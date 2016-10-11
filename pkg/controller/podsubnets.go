package controller

import (
	"github.com/golang/glog"
	"k8s.io/client-go/1.4/pkg/api"

	"github.com/sapcc/kube-parrot/pkg/bgp"
)

type PodSubnetsController struct {
	client *clientset.Clientset

	store      cache.Store
	controller *framework.Controller
	bgp        *bgp.Server
}

func NewPodSubnetsController(client *clientset.Clientset, bgp *bgp.Server) *PodSubnetsController {
	n := &PodSubnetsController{
		client: client,
		bgp:    bgp,
	}

	n.store, n.controller = framework.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return n.client.Core().Nodes().List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return n.client.Core().Nodes().Watch(options)
			},
		},
		&api.Node{},
		controller.NoResyncPeriodFunc(),
		framework.ResourceEventHandlerFuncs{
			AddFunc:    n.addNode,
			DeleteFunc: n.deleteNode,
		},
	)

	return n
}

func (n *PodSubnetsController) Run(stopCh <-chan struct{}) {
	n.controller.Run(stopCh)
}

func (n *PodSubnetsController) addNode(obj interface{}) {
	node := obj.(*api.Node)
	glog.V(3).Infof("Node created: %s", node.GetName())

	route, err := getPodSubnetRoute(node)
	if err != nil {
		glog.Warningf("Couldn't add pod subnet for %s: %s", node.GetName(), err)
		return
	}

	n.bgp.AddPath(route)
}

func (n *PodSubnetsController) deleteNode(obj interface{}) {
	node := obj.(*api.Node)
	glog.V(3).Infof("Node deleted: %s", node.GetName())

	route, err := getPodSubnetRoute(node)
	if err != nil {
		glog.Warningf("Couldn't add pod subnet for %s: %s", node.GetName(), err)
		return
	}

	n.bgp.DeletePath(route)
}
