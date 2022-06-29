package util

import (
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"sort"
)

type Nacos struct {
	NacosNameSpace string
	NacosGroupName string
	DevEnv         bool
	cc             constant.ClientConfig
	sc             []constant.ServerConfig
	nmClient       naming_client.INamingClient
}

func NewNacos(nacosNameSpace string) *Nacos {
	nacos := &Nacos{}
	nacos.NacosNameSpace = nacosNameSpace
	nacos.NacosGroupName = "DEFAULT_GROUP"
	nacos.InitNacos()
	return nacos
}

func (nacos *Nacos) ResgisterInstance(serviceName string, env int, ip string, port int) (bool, error) {
	metaData := map[string]string{}
	if env == ENV_GRAY {
		metaData["env"] = "grayTag"
	}
	return nacos.nmClient.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: serviceName,
		ClusterName: "cluster-live",
		GroupName:   nacos.NacosGroupName,
		Ip:          ip,
		Port:        uint64(port),
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    metaData,
	})
}

func (nacos *Nacos) getNacosDomain() string {
	return "nacos-live.aijidou.com"
}

func (nacos *Nacos) InitNacos() {
	nacos.sc = []constant.ServerConfig{
		{
			IpAddr:      nacos.getNacosDomain(),
			Port:        8848,
			ContextPath: "/nacos",
			Scheme:      "http",
		},
	}

	//create ClientConfig
	namespaceId := nacos.NacosNameSpace
	nacos.cc = constant.ClientConfig{
		NamespaceId:         namespaceId, //namespace id
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// create naming client
	var err error
	nacos.nmClient, err = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &nacos.cc,
			ServerConfigs: nacos.sc,
		},
	)

	if err != nil {
		panic(err)
	}
}

const (
	ENV_LIVE = 0
	ENV_GRAY = 1
	ENV_ALL  = 2
)

func (nacos *Nacos) GetService(env int, serviceName string) (model.Service, error) {
	srv, err := nacos.nmClient.GetService(vo.GetServiceParam{
		Clusters:    []string{"cluster-live"},
		ServiceName: serviceName,
		GroupName:   nacos.NacosGroupName,
	})
	if env == ENV_ALL {
		return srv, err
	}

	targetHosts := []model.Instance{}
	for _, item := range srv.Hosts {
		if env == ENV_GRAY {
			if item.Metadata["env"] == "grayTag" {
				targetHosts = append(targetHosts, item)
			}
		} else if env == ENV_LIVE {
			if item.Metadata["env"] == "" {
				targetHosts = append(targetHosts, item)
			}
		}
	}

	srv.Hosts = targetHosts
	return srv, err
}

type ServiceInstances []model.Instance

func (si *ServiceInstances) Len() int {
	return len(*si)
}

func (si *ServiceInstances) Less(i, j int) bool {
	return (*si)[i].Weight < (*si)[j].Weight
}

func (si *ServiceInstances) Swap(i, j int) {
	(*si)[i], (*si)[j] = (*si)[j], (*si)[i]
}

func (nacos *Nacos) GetTargetMachine(targetService *model.Service) (string, uint64) {
	targetHosts := ServiceInstances(targetService.Hosts)
	sort.Sort(&targetHosts)

	weights := []float64{}
	liveHosts := ServiceInstances{}
	for _, target := range targetHosts {
		if target.Enable && target.Healthy {
			liveHosts = append(liveHosts, target)
			weights = append(weights, target.Weight)
		}
	}
	if len(liveHosts) <= 0 {
		return "", 0
	}

	ipNo := GetWeight(weights)
	return liveHosts[ipNo].Ip, liveHosts[ipNo].Port
}
