package client

const (
	DefaultHost   = "localhost"
	DefaultPort   = 8089
	DefaultScheme = "https"
)

const (
	//pathDeploymentServers = "deployment/server/"
	//path_inputs = "data/inputs/"
	//pathModular_inputs = "data/modular-inputs"
	pathUsers            = "authentication/users/"
	pathStoragePasswords = "storage/passwords"
	pathInfo             = "/services/server/info"
)

type SplunkService struct {
	Host  string
	Port  int
	Debug bool
	authUser
	authToken
}

func (ss *SplunkService) get(pathSegment string, owner string, app string, sharing string) /* **query*/ {
	return
}

func (ss *SplunkService) Info() {
	return
}

func (ss *SplunkService) StoragePasswords() {
	return
}
