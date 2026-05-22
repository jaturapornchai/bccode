package timezone

type Timezone struct {
	TimezoneLabel string `json:"timezone_label" bson:"timezone_label"`
	TimezoneOffset string `json:"timezone_offset" bson:"timezone_offset"`
}
