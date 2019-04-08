package kubernetes

import (
	"fmt"
	v12 "github.com/ericchiang/k8s/apis/apps/v1"
	"github.com/ericchiang/k8s/apis/core/v1"
	v13 "github.com/ericchiang/k8s/apis/rbac/v1"
	"github.com/ghodss/yaml"
	"github.com/infinityworks/fk-infra/util"
	"strings"
)

const fluentBitTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  name: logging

---

apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: fluent-bit-logging
    kubernetes.io/cluster-service: "true"
    version: v1
  name: fluent-bit
  namespace: logging
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      name: fluent-bit-logging
  template:
    metadata:
      annotations:
        prometheus.io/path: /api/v1/metrics/prometheus
        prometheus.io/port: "2020"
        prometheus.io/scrape: "true"
      creationTimestamp: null
      labels:
        k8s-app: fluent-bit-logging
        kubernetes.io/cluster-service: "true"
        name: fluent-bit-logging
        version: v1
    spec:
      containers:
      - image: fluent/fluent-bit:1.0.5
        imagePullPolicy: Always
        name: fluent-bit
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/log
          name: varlog
        - mountPath: /var/log/containers
          name: varlogcontainers
          readOnly: true
        - mountPath: /var/lib/docker/containers
          name: varlibdockercontainers
        - mountPath: /fluent-bit/etc/
          name: fluent-bit-config
      - args:
        - -target
        - http://logging
        image: cllunsford/aws-signing-proxy:latest
        imagePullPolicy: Always
        name: aws-signing-proxy
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: fluent-bit
      serviceAccountName: fluent-bit
      terminationGracePeriodSeconds: 10
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Exists
      - effect: NoExecute
        operator: Exists
      - effect: NoSchedule
        operator: Exists
      volumes:
      - name: varlog
        volumeSource:
          hostPath:
            path: /var/log
      - name: varlogcontainers
        volumeSource:
          hostPath:
            path: /var/log/containers
      - name: varlibdockercontainers
        volumeSource:
          hostPath:
            path: /var/lib/docker/containers
      - name: fluent-bit-config
        volumeSource:
          configMap:
            localObjectReference:
              name: fluent-bit-config

---

apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: logging
  labels:
    k8s-app: fluent-bit
data:
  # Configuration files: server, input, filters and output
  # ======================================================
  fluent-bit.conf: |
    [SERVICE]
        Flush         1
        Log_Level     info
        Daemon        off
        Parsers_File  parsers.conf
        HTTP_Server   On
        HTTP_Listen   0.0.0.0
        HTTP_Port     2020

    @INCLUDE input-kubernetes.conf
    @INCLUDE filter-kubernetes.conf
    @INCLUDE output-elasticsearch.conf

  input-kubernetes.conf: |
    [INPUT]
        Name              tail
        Tag               kube.*
        Path              /var/log/containers/*.log
        Parser            docker
        DB                /var/log/flb_kube.db
        Mem_Buf_Limit     5MB
        Skip_Long_Lines   On
        Refresh_Interval  10

  filter-kubernetes.conf: |
    [FILTER]
        Name                kubernetes
        Match               kube.*
        Merge_Log           On
        K8S-Logging.Parser  On

  output-elasticsearch.conf: |
    [OUTPUT]
        Name            es
        Match           *
        Host            127.0.0.1
        Port            8080
        Logstash_Format On
        Replace_Dots    On
        Retry_Limit     False

  parsers.conf: |
    [PARSER]
        Name   apache
        Format regex
        Regex  ^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)")?$
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z

    [PARSER]
        Name   apache2
        Format regex
        Regex  ^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^ ]*) +\S*)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)")?$
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z

    [PARSER]
        Name   apache_error
        Format regex
        Regex  ^\[[^ ]* (?<time>[^\]]*)\] \[(?<level>[^\]]*)\](?: \[pid (?<pid>[^\]]*)\])?( \[client (?<client>[^\]]*)\])? (?<message>.*)$

    [PARSER]
        Name   nginx
        Format regex
        Regex ^(?<remote>[^ ]*) (?<host>[^ ]*) (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)")?$
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z

    [PARSER]
        Name   json
        Format json
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z

    [PARSER]
        Name        docker
        Format      json
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%S.%L
        Time_Keep   On
        # Command      |  Decoder | Field | Optional Action
        # =============|==================|=================
        Decode_Field_As   escaped    log

    [PARSER]
        Name        syslog
        Format      regex
        Regex       ^\<(?<pri>[0-9]+)\>(?<time>[^ ]* {1,2}[^ ]* [^ ]*) (?<host>[^ ]*) (?<ident>[a-zA-Z0-9_\/\.\-]*)(?:\[(?<pid>[0-9]+)\])?(?:[^\:]*\:)? *(?<message>.*)$
        Time_Key    time
        Time_Format %b %d %H:%M:%S

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: fluent-bit
  namespace: logging

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: fluent-bit-read
rules:
- apiGroups: [""]
  resources:
  - namespaces
  - pods
  verbs: ["get", "list", "watch"]

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: fluent-bit-read
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: fluent-bit-read
subjects:
- kind: ServiceAccount
  name: fluent-bit
  namespace: logging
`

func ApplyFluentBitLogging(elasticsearchEndpoint, region string) {
	documentItems := strings.Split(fluentBitTemplate, "---")

	var namespace v1.Namespace
	var daemonSet v12.DaemonSet
	var configMap v1.ConfigMap
	var serviceAccount v1.ServiceAccount
	var clusterRole v13.ClusterRole
	var clusterRoleBinding v13.ClusterRoleBinding

	util.CheckError(yaml.Unmarshal([]byte(documentItems[0]), &namespace))
	util.CheckError(yaml.Unmarshal([]byte(documentItems[1]), &daemonSet))
	util.CheckError(yaml.Unmarshal([]byte(documentItems[2]), &configMap))
	util.CheckError(yaml.Unmarshal([]byte(documentItems[3]), &serviceAccount))
	util.CheckError(yaml.Unmarshal([]byte(documentItems[4]), &clusterRole))
	util.CheckError(yaml.Unmarshal([]byte(documentItems[5]), &clusterRoleBinding))

	daemonSet.Spec.Template.Spec.Containers[1].Args[1] = fmt.Sprintf("https://%s", elasticsearchEndpoint)
	daemonSet.Spec.Template.Spec.Containers[1].Env = []*v1.EnvVar{{Name: util.String("AWS_REGION"), Value: util.String(region)}}

	CreateOrUpdate(&namespace)
	CreateOrUpdate(&daemonSet)
	CreateOrUpdate(&configMap)
	CreateOrUpdate(&serviceAccount)
	CreateOrUpdate(&clusterRole)
	CreateOrUpdate(&clusterRoleBinding)
}
