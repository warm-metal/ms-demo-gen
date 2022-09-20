package manifest

const deployTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: "{{.Name}}"
  namespace: "{{.Namespace}}"
  labels:
    app: "{{.App}}"
    version: "{{.Version}}"
    origin: msdgen
spec:
  replicas: {{.NumReplicas}}
  selector:
    matchLabels:
      app: "{{.App}}"
      version: "{{.Version}}"
  template:
    metadata:
      labels:
        app: "{{.App}}"
        version: "{{.Version}}"
        origin: msdgen
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
        - name: ENV_DISCARD_UPSTREAM_PAYLOAD
          value: "1"
{{- else}}
{{- if .HasResourceConstraints}}
        resources:
          requests:
            cpu: "{{.CPURequest}}"
          limits:
            cpu: "{{.CPULimit}}"
{{- end}}
        ports:
        - containerPort: 80
{{- end}}
---
`
const serviceTemplate = `apiVersion: v1
kind: Service
metadata:
  name: "{{.App}}"
  namespace: "{{.Namespace}}"
  labels:
    app: "{{.App}}"
    origin: "msdgen"
spec:
  selector:
    app: "{{.App}}"
  ports:
    - protocol: TCP
      appProtocol: http
      port: 80
      targetPort: 80
      name: "http-80"
---
`
