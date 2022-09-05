package manifest

const deployTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    app: {{.App}}
    origin: msdgen
spec:
  replicas: {{.NumReplicas}}
  selector:
    matchLabels:
      app: {{.App}}
      svc: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.App}}
        svc: {{.Name}}
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
{{- else}}
{{- if .HasResourceConstraints}}
        resources:
{{- if .CPURequest}}
          requests:
            cpu: "{{.CPURequest}}"
{{- end}}
{{- if .CPULimit}}
          limits:
            cpu: "{{.CPULimit}}"
{{- end}}
{{- end}}
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
    app: {{.App}}
    origin: msdgen
spec:
  selector:
    svc: {{.Name}}
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
---
`
