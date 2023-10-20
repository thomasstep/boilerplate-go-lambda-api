import * as fs from 'fs';
import * as path from 'path';

import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as sns from 'aws-cdk-lib/aws-sns';
import * as snsSub from 'aws-cdk-lib/aws-sns-subscriptions';
import * as sqs from 'aws-cdk-lib/aws-sqs';

const srcDirectory = '../../src';
const ddbEnvVarName = 'PRIMARY_TABLE_NAME';
const snsEnvVarName = 'PRIMARY_SNS_TOPIC_ARN';

function connectDdbToLambdas(table: dynamodb.Table, lambdas: lambda.Function[], envVarName: string) {
  lambdas.forEach((lambda) => {
    table.grantFullAccess(lambda);
    lambda.addEnvironment(envVarName, table.tableName);
  })
}

interface IApiProps extends cdk.StackProps {
  primaryTable: dynamodb.Table,
  // snsTopic: sns.Topic,
  configFile: string,
}

export class Api extends cdk.Stack {
  public api: apigateway.RestApi;

  constructor(scope: Construct, id: string, props: IApiProps) {
    super(scope, id, props);

    const {
      configFile,
      primaryTable,
      // snsTopic,
    } = props;

    const filePath = path.join(process.cwd(), configFile);
    const contents = fs.readFileSync(filePath, 'utf8');
    const config = JSON.parse(contents);
    const {
      configItem,
      eventOperations,
    } = config;

    function baseLambdaConfig(target: string) {
      return {
        handler: 'main', // Because the build output is called main
        runtime: lambda.Runtime.PROVIDED_AL2,
        logRetention: logs.RetentionDays.ONE_WEEK,
        code: lambda.Code.fromAsset(path.join(__dirname, srcDirectory), {
          bundling: {
            image: cdk.DockerImage.fromRegistry('golang:1.21.3'),
            // user: "root",
            command: [
              'bash', '-c',
              `GOCACHE=/tmp go mod tidy && GOCACHE=/tmp GOOS=linux GOARCH=amd64 go build -o /asset-output/bootstrap ./cmd/${target}`
            ]
          },
        }),
        environment: {
          CORS_ALLOW_ORIGIN_HEADER: config.corsAllowOriginHeader,
        },
        timeout: cdk.Duration.seconds(5),
      }
    }

    const authorizerLambda = new lambda.Function(this, 'request-authorizer-lambda', {
        ...baseLambdaConfig('lambdaAuthorizer'),
    });
    // Use if JWT authorizer
    // authorizerLambda.addEnvironment('JWKS_URL', jwksUrl);
    // authorizerLambda.addEnvironment('JWK_ID', jwkId);
    const authorizer = new apigateway.RequestAuthorizer(
      this,
      'request-authorizer',
      {
        handler: authorizerLambda,
        resultsCacheTtl: cdk.Duration.seconds(3600),
        identitySources: [apigateway.IdentitySource.header('Authorization')]
      },
    );

    // API Gateway log group
    const gatewayLogGroup = new logs.LogGroup(this, 'api-access-logs', {
      retention: logs.RetentionDays.ONE_WEEK,
    });

    // The API Gateway itself (same name as stack)
    const restApi = new apigateway.RestApi(this, id, {
      deploy: true,
      deployOptions: {
        loggingLevel: apigateway.MethodLoggingLevel.ERROR,
        accessLogDestination: new apigateway.LogGroupLogDestination(gatewayLogGroup),
      },
      endpointTypes: [apigateway.EndpointType.REGIONAL],
      defaultCorsPreflightOptions: {
        allowOrigins: ['*'],
        allowCredentials: true,
      },
      apiKeySourceType: apigateway.ApiKeySourceType.HEADER,
    });
    this.api = restApi;

    // API key
    const apiKey = restApi.addApiKey('api-key');
    const usagePlan = new apigateway.UsagePlan(this, 'usage-plan', {
      throttle: {
        burstLimit: 5000,
        rateLimit: 10000,
      },
      apiStages: [
        {
          api: restApi,
          stage: restApi.deploymentStage,
        },
      ],
    });
    usagePlan.addApiKey(apiKey);

    // Adding this method otherwise there's an error that the authorizer isn't attached to an API Gateway
    restApi.root.addMethod('GET', undefined, {
      authorizer,
      authorizationType: apigateway.AuthorizationType.CUSTOM,
    });

    // This sends CORS headers with lambda authorizer response
    new apigateway.GatewayResponse(this, 'default-4xx-gateway-response', {
      restApi: restApi,
      type: apigateway.ResponseType.DEFAULT_4XX,
      responseHeaders: {
        'method.response.header.Access-Control-Allow-Origin': `'${config.corsAllowOriginHeader}'`,
        'method.response.header.Access-Control-Allow-Credentials': "'true'",
      },
    });
    new apigateway.GatewayResponse(this, 'default-5xx-gateway-response', {
      restApi: restApi,
      type: apigateway.ResponseType.DEFAULT_5XX,
      responseHeaders: {
        'method.response.header.Access-Control-Allow-Origin': `'${config.corsAllowOriginHeader}'`,
        'method.response.header.Access-Control-Allow-Credentials': "'true'",
      },
    });

    // ************************************************************************
    // Build models
    // ************************************************************************

    // Request validators
    const validateBodyValidator = restApi.addRequestValidator('validateBody', {
      requestValidatorName: 'validateBody',
      validateRequestBody: true,
    });
    const validateParamsValidator = restApi.addRequestValidator('validateParams', {
      requestValidatorName: 'validateParams',
      validateRequestParameters: true,
    });

    // ************************************************************************
    // Build Lambdas and their methods
    // ************************************************************************
    const createLambda = new lambda.Function(this, 'create', baseLambdaConfig('create'));
    const readLambda = new lambda.Function(this, 'read', baseLambdaConfig('read'));
    const updateLambda = new lambda.Function(this, 'update', baseLambdaConfig('update'));
    const deleteLambda = new lambda.Function(this, 'delete', baseLambdaConfig('delete'));

    // Uncomment if there are shared environment variables that need to be set
    // [
    //   createLambda,
    //   readLambda,
    //   updateLambda,
    //   deleteLambda,
    // ].forEach((lambda) => {
    //   lambda.addEnvironment(sharedEnvVarName, value);
    // });

    // Uncomment if there are one-off environment variables that need to be set
    // createLambda.addEnvironment(oneOffEnvVarName, value);

    connectDdbToLambdas(
      primaryTable,
      [
        createLambda,
        readLambda,
        updateLambda,
        deleteLambda,
      ],
      ddbEnvVarName,
    );

    // ************************************************************************
    // Build API paths
    // ************************************************************************

    const v1Resource = restApi.root.addResource('v1');
    const entityResource = v1Resource.addResource('entity');
    const entityIdResource = entityResource.addResource('{entityId}');

    // ************************************************************************
    // Add methods
    // ************************************************************************

    entityResource.addMethod(
      'POST',
      new apigateway.LambdaIntegration(createLambda, {}),
      {
        authorizationType: apigateway.AuthorizationType.CUSTOM,
        authorizer,
        // requestValidator: validateBodyValidator,
        // requestModels: {
        //   'application/json': createModel,
        // },
      },
    );

    entityIdResource.addMethod(
      'DELETE',
      new apigateway.LambdaIntegration(deleteLambda, {}),
      {
        authorizationType: apigateway.AuthorizationType.CUSTOM,
        authorizer,
      },
    );

    entityIdResource.addMethod(
      'GET',
      new apigateway.LambdaIntegration(readLambda, {}),
      {
        authorizationType: apigateway.AuthorizationType.CUSTOM,
        authorizer,
      },
    );

    entityIdResource.addMethod(
      'PUT',
      new apigateway.LambdaIntegration(updateLambda, {}),
      {
        authorizationType: apigateway.AuthorizationType.CUSTOM,
        authorizer,
        // requestValidator: validateBodyValidator,
        // requestModels: {
        //   'application/json': updateModel,
        // },
      },
    );

    // *************************************************************************
    // Create async Lambdas and connect to SNS
    // *************************************************************************

    // const asyncLambdaNames = [
    //   {
    //     camelCase: 'eventAction',
    //     kebabCase: 'event-action',
    //     operations: [
    //       eventOperations.eventActionEvent,
    //     ],
    //     usesDb: true,
    //   },
    // ];

    // asyncLambdaNames.forEach((config) => {
    //   // Add alarms if any of these fail
    //   const dlq = new sqs.Queue(this, `${config.kebabCase}-dlq`, {});
    //   const lambdaFunction = new lambda.Function(
    //     this,
    //     `${config.kebabCase}-lambda`,
    //     {
    //       ...baseLambdaConfig(config.camelCase),
    //       timeout: cdk.Duration.seconds(20), // Giving background tasks a longer timeout
    //       deadLetterQueue: dlq,
    //     },
    //   );
    //   snsTopic.addSubscription(new snsSub.LambdaSubscription(
    //     lambdaFunction,
    //     {
    //       filterPolicy: {
    //         operation: sns.SubscriptionFilter.stringFilter({
    //           allowlist: config.operations,
    //         }),
    //       },
    //     }
    //   ));
    //   if (config.usesDb) {
    //     primaryTable.grantFullAccess(lambdaFunction);
    //     lambdaFunction.addEnvironment(ddbEnvVarName, primaryTable.tableName);
    //   }
    // });

    // *************************************************************************
    // Setup Lambdas that publish to SNS
    // *************************************************************************

    // const lambdasThatPublish = [
    //   createLambda,
    //   updateLambda,
    //   deleteLambda,
    // ];
    // lambdasThatPublish.forEach((lambda) => {
    //   snsTopic.grantPublish(lambda);
    //   lambda.addEnvironment(snsEnvVarName, snsTopic.topicArn);
    // });
  }
}
