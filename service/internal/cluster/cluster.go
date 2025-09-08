package cluster

var (
	GlobalCluster *Cluster
)

type Cluster struct {
}

func (c *Cluster) GetOtherHost() []string {
	return []string{}
}

func NewClusterFromConfig() *Cluster {

	return &Cluster{}
}

func init() {
	GlobalCluster = NewClusterFromConfig()
}
