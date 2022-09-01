package manifest

const deployTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    app: {{.Name}}
    origin: wsd
spec:
  replicas: {{.NumReplicas}}
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
        origin: wsd
    spec:
      containers:
      - name: {{.Name}}
        image: {{.Image}}
        env:
        - name: ENV_PAYLOAD_SIZE
          value: "{{.PayloadSize}}"
        - name: ENV_UPLOAD_SIZE
          value: "{{.UploadSize}}"
        - name: ENV_UPSTREAM
          value: "{{.JoinUpstreams}}"
        - name: ENV_QUERY_IN_PARALLEL
          value: "{{.QueryInParallelInInt}}"
        - name: ENV_USE_LONG_CONNECTION
          value: "{{.LongConnInInt}}"
        - name: ENV_TIMEOUT
          value: "{{.Timeout}}"
{{- if not .Address}}
        - name: ENV_CONCURRENT_PROCS
          value: "{{.NumConcurrentProc}}"
        - name: ENV_INTERVAL_BETWEEN_QUERIES
          value: "{{.QueryInterval}}"
{{- else}}
        ports:
        - containerPort: 80
{{- end}}
---
`
const serviceTemplate = `apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    app: {{.Name}}
    origin: wsd
spec:
  selector:
    app: {{.Name}}
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
---
`