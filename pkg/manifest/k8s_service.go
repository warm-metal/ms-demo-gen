package manifest

const workloadTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    app: {{.Name}}
spec:
  replicas: {{.NumReplicas}}
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
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
		  value: "{{.QueryInParallel}}"
		- name: ENV_USE_LONG_CONNECTION
		  value: "{{.LongConn}}"
		- name: ENV_TIMEOUT
		  value: "{{.Timeout}}"
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  selector:
    app: {{.Name}}
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
---
`
