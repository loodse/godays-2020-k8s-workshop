module github.com/loodse/godays-2020-k8s-workshop/smart-home

go 1.13

require (
	github.com/gizak/termui/v3 v3.0.0-00010101000000-000000000000
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.6.0
	github.com/onsi/gomega v1.4.2
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.2
)

replace github.com/gizak/termui/v3 => github.com/thetechnick/termui/v3 v3.1.1-0.20200116102044-73b5d38aef80
