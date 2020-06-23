module github.com/ontology-community/onRobot

go 1.13

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/hashicorp/golang-lru v0.5.3
	github.com/jinzhu/gorm v1.9.12
	github.com/ontio/ontology v0.0.0-20200622110714-6712611a8631
	github.com/ontio/ontology-crypto v1.0.9
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/scylladb/go-set v1.0.2
)

replace github.com/ontio/ontology v0.0.0-20200622110714-6712611a8631 => ../../laizy/ontology
