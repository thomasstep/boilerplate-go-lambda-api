package types

type Entity struct {
	Id        string `json:"id" dynamodbav:"-"`
	Name      string `json:"name,omitempty" dynamodbav:"name"` // Optional
}

type EntityList struct {
	Entities      []Entity     `json:"entity"`
	Pagination Pagination `json:"pagination"`
}

type EntityUpdates struct {
	Name string `json:"name" dynamodbav:"name"`
}

type DdbEntityItem struct {
	Entity
	Id          string `dynamodbav:"id"`
	SecondaryId string `dynamodbav:"secondaryId"`
	CreatedTime string `json:"createdTime" dynamodbav:"createdTime"`
	UpdatedTime string `json:"updatedTime" dynamodbav:"updatedTime"`
}

type Pagination struct {
	NextToken *string `json:"nextToken"`
}

type DdbPrimaryKey struct {
	Id          string `dynamodbav:"id"`
	SecondaryId string `dynamodbav:"secondaryId"`
}
