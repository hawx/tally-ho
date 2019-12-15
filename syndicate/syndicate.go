package syndicate

type Syndicator interface {
	Create(data map[string][]interface{}) (location string, err error)
	Config() Config
}

type Config struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}
