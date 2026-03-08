package timezone

type Timezone struct {
	TimezoneLabel  string `json:"timezonelabel" bson:"timezonelabel"`
	TimezoneOffset string `json:"timezoneoffset" bson:"timezoneoffset"`
}
