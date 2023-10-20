#!/usr/bin/env node
import * as fs from 'fs';
import * as path from 'path';

import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { Tables } from '../lib/tables';
import { Api } from '../lib/api';

const devFilePath = path.join(process.cwd(), 'config-dev.json');
const devContents = fs.readFileSync(devFilePath, 'utf8');
const devConfig = JSON.parse(devContents);

const app = new cdk.App();

const tables = new Tables(app, 'tables', {
  env: devConfig.cdkEnvironment,
});
new Api(app, 'api', {
  configFile: 'config-dev.json',
  env: devConfig.cdkEnvironment,
  primaryTable: tables.primaryTable,
});
