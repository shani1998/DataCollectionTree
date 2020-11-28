## deploy
``` shell
git clone https://github.com/shani1998/DataCollectionTree.git
kubectl create -f DataCollectionTree/data-collection-deploy.yaml
kubectl get svc data-collection-tree -o yaml >>/tmp/svc.yaml
clusterIP=`sed -n '/spec:/,/status:/p' /tmp/svc.yaml | grep -oP '(?<=clusterIP: ).*'`
```

## construct tree
```shell
curl -i -X POST -H "Content-Type: application/json" -d '{"dim": [{ "key": "device", "val": "mobile" }, { "key": "country", "val": "US"}],"metrics": [{	"key": "webreq","val": 80},{"key": "timespent","val": 70}]}' http://$clusterIP:8080/v1/insert
curl -i -X POST -H "Content-Type: application/json" -d '{"dim": [{ "key": "device", "val": "web" },    { "key": "country", "val": "US"}],"metrics": [{	"key": "webreq","val": 110},{"key": "timespent","val": 60}]}' http://$clusterIP:8080/v1/insert
curl -i -X POST -H "Content-Type: application/json" -d '{"dim": [{ "key": "device", "val": "tablet" }, { "key": "country", "val": "US"}],"metrics": [{	"key": "webreq","val": 30},{"key": "timespent","val": 50}]}' http://$clusterIP:8080/v1/insert
curl -i -X POST -H "Content-Type: application/json" -d '{"dim": [{ "key": "device", "val": "mobile" }, { "key": "country", "val": "IN"}],"metrics": [{	"key": "webreq","val": 70},{"key": "timespent","val": 30}]}' http://$clusterIP:8080/v1/insert
curl -i -X POST -H "Content-Type: application/json" -d '{"dim": [{ "key": "device", "val": "web" },    { "key": "country", "val": "IN"}],"metrics": [{	"key": "webreq","val": 50},{"key": "timespent","val": 50}]}' http://$clusterIP:8080/v1/insert
```

## query data
```shell
curl -i -X GET -H "Content-Type: application/json" -d '{"dim": [{ "key": "country","val": "IN"} ]}' http://$clusterIP:8080/v1/query
curl -i -X GET -H "Content-Type: application/json" -d '{"dim": [{ "key": "country","val": "US"} ]}' http://$clusterIP:8080/v1/query
curl -i -X GET -H "Content-Type: application/json" -d '{"dim": [{ "key": "country","val": "NL"} ]}' http://$clusterIP:8080/v1/query
```
