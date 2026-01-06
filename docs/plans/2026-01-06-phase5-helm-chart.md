# Phase 5: Helm Chart - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a Helm chart for deploying the central API, Ollama, and optionally Postgres to Kubernetes.

**Architecture:** Standard Helm chart with configurable storage backend (Postgres internal or MongoDB external), network policies, probes, and init containers for migrations.

**Tech Stack:** Helm 3, Kubernetes 1.25+

**Prerequisites:** Phase 4 complete (API server and shim exist)

---

## Task 1: Create Chart Scaffolding

**Files:**
- Create: `charts/engram-cogitator/Chart.yaml`
- Create: `charts/engram-cogitator/values.yaml`
- Create: `charts/engram-cogitator/.helmignore`

**Step 1: Create Chart.yaml**

```yaml
# charts/engram-cogitator/Chart.yaml
apiVersion: v2
name: engram-cogitator
description: Persistent semantic memory for Claude Code teams
type: application
version: 0.1.0
appVersion: "1.0.0"
keywords:
  - claude
  - mcp
  - memory
  - ai
maintainers:
  - name: MereWhiplash
    url: https://github.com/MereWhiplash
sources:
  - https://github.com/MereWhiplash/engram-cogitator
```

**Step 2: Create values.yaml**

```yaml
# charts/engram-cogitator/values.yaml

# Image configuration
image:
  repository: ghcr.io/merewhiplash/engram-cogitator
  tag: ""  # Defaults to appVersion
  pullPolicy: IfNotPresent

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

# API server configuration
api:
  replicas: 2
  resources:
    requests:
      memory: 256Mi
      cpu: 250m
    limits:
      memory: 512Mi
      cpu: 500m
  nodeSelector: {}
  tolerations: []
  affinity: {}

# Storage configuration
storage:
  # Driver: "postgres" or "mongodb"
  driver: postgres

  postgres:
    # Use internal StatefulSet or external
    internal: true
    # External connection (used if internal: false)
    host: ""
    port: 5432
    database: engram
    username: engram
    # Password from secret
    existingSecret: ""
    existingSecretKey: password
    # If no existingSecret, create one with this password
    password: ""
    # Internal StatefulSet settings
    resources:
      requests:
        memory: 512Mi
        cpu: 250m
      limits:
        memory: 1Gi
        cpu: 500m
    persistence:
      size: 10Gi
      storageClass: ""

  mongodb:
    # User provides URI (Atlas or self-hosted)
    uri: ""
    database: engram
    # URI from secret
    existingSecret: ""
    existingSecretKey: uri

# Ollama configuration
ollama:
  enabled: true
  model: nomic-embed-text
  resources:
    requests:
      memory: 2Gi
      cpu: 1
    limits:
      memory: 4Gi
      cpu: 2
  persistence:
    size: 10Gi
    storageClass: ""
  nodeSelector: {}
  tolerations: []

# Ingress configuration
ingress:
  enabled: false
  className: ""
  annotations: {}
  hosts:
    - host: engram.local
      paths:
        - path: /
          pathType: Prefix
  tls: []

# Network policies
networkPolicy:
  enabled: true

# Pod disruption budget
podDisruptionBudget:
  enabled: true
  minAvailable: 1

# Horizontal pod autoscaler
autoscaling:
  enabled: false
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80

# Service account
serviceAccount:
  create: true
  annotations: {}
  name: ""
```

**Step 3: Create .helmignore**

```
# charts/engram-cogitator/.helmignore
.git
.gitignore
*.md
.helmignore
```

**Step 4: Verify structure**

Run: `ls -la charts/engram-cogitator/`
Expected: Chart.yaml, values.yaml, .helmignore

**Step 5: Commit**

```bash
git add charts/
git commit -m "feat(helm): add chart scaffolding with values"
```

---

## Task 2: Create Template Helpers

**Files:**
- Create: `charts/engram-cogitator/templates/_helpers.tpl`

**Step 1: Write helpers**

```yaml
{{/* charts/engram-cogitator/templates/_helpers.tpl */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "engram-cogitator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "engram-cogitator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "engram-cogitator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "engram-cogitator.labels" -}}
helm.sh/chart: {{ include "engram-cogitator.chart" . }}
{{ include "engram-cogitator.selectorLabels" . }}
app.kubernetes.io/version: {{ .Values.image.tag | default .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "engram-cogitator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "engram-cogitator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "engram-cogitator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "engram-cogitator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
API component name
*/}}
{{- define "engram-cogitator.api.name" -}}
{{- printf "%s-api" (include "engram-cogitator.fullname" .) }}
{{- end }}

{{/*
Ollama component name
*/}}
{{- define "engram-cogitator.ollama.name" -}}
{{- printf "%s-ollama" (include "engram-cogitator.fullname" .) }}
{{- end }}

{{/*
Postgres component name
*/}}
{{- define "engram-cogitator.postgres.name" -}}
{{- printf "%s-postgres" (include "engram-cogitator.fullname" .) }}
{{- end }}

{{/*
Postgres DSN
*/}}
{{- define "engram-cogitator.postgres.dsn" -}}
{{- if .Values.storage.postgres.internal }}
{{- printf "postgres://%s:$(POSTGRES_PASSWORD)@%s:5432/%s?sslmode=disable" .Values.storage.postgres.username (include "engram-cogitator.postgres.name" .) .Values.storage.postgres.database }}
{{- else }}
{{- printf "postgres://%s:$(POSTGRES_PASSWORD)@%s:%d/%s?sslmode=disable" .Values.storage.postgres.username .Values.storage.postgres.host (.Values.storage.postgres.port | int) .Values.storage.postgres.database }}
{{- end }}
{{- end }}

{{/*
Ollama URL
*/}}
{{- define "engram-cogitator.ollama.url" -}}
{{- printf "http://%s:11434" (include "engram-cogitator.ollama.name" .) }}
{{- end }}
```

**Step 2: Commit**

```bash
git add charts/engram-cogitator/templates/_helpers.tpl
git commit -m "feat(helm): add template helpers"
```

---

## Task 3: Create API Deployment and Service

**Files:**
- Create: `charts/engram-cogitator/templates/api-deployment.yaml`
- Create: `charts/engram-cogitator/templates/api-service.yaml`

**Step 1: Write API deployment**

```yaml
# charts/engram-cogitator/templates/api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "engram-cogitator.api.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
    app.kubernetes.io/component: api
spec:
  replicas: {{ .Values.api.replicas }}
  selector:
    matchLabels:
      {{- include "engram-cogitator.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: api
  template:
    metadata:
      labels:
        {{- include "engram-cogitator.selectorLabels" . | nindent 8 }}
        app.kubernetes.io/component: api
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "engram-cogitator.serviceAccountName" . }}
      {{- if eq .Values.storage.driver "postgres" }}
      initContainers:
        - name: wait-for-postgres
          image: busybox:1.36
          command: ['sh', '-c', 'until nc -z {{ include "engram-cogitator.postgres.name" . }} 5432; do echo waiting for postgres; sleep 2; done']
        - name: migrate
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          command: ["/ec-api", "migrate"]
          env:
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.storage.postgres.existingSecret | default (printf "%s-postgres" (include "engram-cogitator.fullname" .)) }}
                  key: {{ .Values.storage.postgres.existingSecretKey | default "password" }}
            - name: DATABASE_URL
              value: {{ include "engram-cogitator.postgres.dsn" . | quote }}
      {{- end }}
      containers:
        - name: api
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          args:
            - --addr=:8080
            - --storage-driver={{ .Values.storage.driver }}
            {{- if eq .Values.storage.driver "postgres" }}
            - --postgres-dsn=$(DATABASE_URL)
            {{- else if eq .Values.storage.driver "mongodb" }}
            - --mongodb-uri=$(MONGODB_URI)
            - --mongodb-database={{ .Values.storage.mongodb.database }}
            {{- end }}
            - --ollama-url={{ include "engram-cogitator.ollama.url" . }}
          env:
            {{- if eq .Values.storage.driver "postgres" }}
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.storage.postgres.existingSecret | default (printf "%s-postgres" (include "engram-cogitator.fullname" .)) }}
                  key: {{ .Values.storage.postgres.existingSecretKey | default "password" }}
            - name: DATABASE_URL
              value: {{ include "engram-cogitator.postgres.dsn" . | quote }}
            {{- else if eq .Values.storage.driver "mongodb" }}
            - name: MONGODB_URI
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.storage.mongodb.existingSecret | default (printf "%s-mongodb" (include "engram-cogitator.fullname" .)) }}
                  key: {{ .Values.storage.mongodb.existingSecretKey | default "uri" }}
            {{- end }}
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
          livenessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 2
            periodSeconds: 5
          resources:
            {{- toYaml .Values.api.resources | nindent 12 }}
      {{- with .Values.api.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.api.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.api.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
```

**Step 2: Write API service**

```yaml
# charts/engram-cogitator/templates/api-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ include "engram-cogitator.api.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
    app.kubernetes.io/component: api
spec:
  type: ClusterIP
  ports:
    - port: 8080
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "engram-cogitator.selectorLabels" . | nindent 4 }}
    app.kubernetes.io/component: api
```

**Step 3: Commit**

```bash
git add charts/engram-cogitator/templates/api-*.yaml
git commit -m "feat(helm): add API deployment and service"
```

---

## Task 4: Create Ollama Deployment and Service

**Files:**
- Create: `charts/engram-cogitator/templates/ollama-deployment.yaml`
- Create: `charts/engram-cogitator/templates/ollama-service.yaml`
- Create: `charts/engram-cogitator/templates/ollama-pvc.yaml`

**Step 1: Write Ollama deployment**

```yaml
# charts/engram-cogitator/templates/ollama-deployment.yaml
{{- if .Values.ollama.enabled }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "engram-cogitator.ollama.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
    app.kubernetes.io/component: ollama
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      {{- include "engram-cogitator.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: ollama
  template:
    metadata:
      labels:
        {{- include "engram-cogitator.selectorLabels" . | nindent 8 }}
        app.kubernetes.io/component: ollama
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      initContainers:
        - name: pull-model
          image: ollama/ollama:latest
          command:
            - sh
            - -c
            - |
              ollama serve &
              sleep 5
              ollama pull {{ .Values.ollama.model }}
              pkill ollama
          volumeMounts:
            - name: ollama-data
              mountPath: /root/.ollama
      containers:
        - name: ollama
          image: ollama/ollama:latest
          ports:
            - name: http
              containerPort: 11434
              protocol: TCP
          livenessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            {{- toYaml .Values.ollama.resources | nindent 12 }}
          volumeMounts:
            - name: ollama-data
              mountPath: /root/.ollama
      volumes:
        - name: ollama-data
          persistentVolumeClaim:
            claimName: {{ include "engram-cogitator.ollama.name" . }}
      {{- with .Values.ollama.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.ollama.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
{{- end }}
```

**Step 2: Write Ollama service**

```yaml
# charts/engram-cogitator/templates/ollama-service.yaml
{{- if .Values.ollama.enabled }}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "engram-cogitator.ollama.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
    app.kubernetes.io/component: ollama
spec:
  type: ClusterIP
  ports:
    - port: 11434
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "engram-cogitator.selectorLabels" . | nindent 4 }}
    app.kubernetes.io/component: ollama
{{- end }}
```

**Step 3: Write Ollama PVC**

```yaml
# charts/engram-cogitator/templates/ollama-pvc.yaml
{{- if .Values.ollama.enabled }}
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ include "engram-cogitator.ollama.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
    app.kubernetes.io/component: ollama
spec:
  accessModes:
    - ReadWriteOnce
  {{- if .Values.ollama.persistence.storageClass }}
  storageClassName: {{ .Values.ollama.persistence.storageClass }}
  {{- end }}
  resources:
    requests:
      storage: {{ .Values.ollama.persistence.size }}
{{- end }}
```

**Step 4: Commit**

```bash
git add charts/engram-cogitator/templates/ollama-*.yaml
git commit -m "feat(helm): add Ollama deployment, service, and PVC"
```

---

## Task 5: Create Postgres StatefulSet (Optional)

**Files:**
- Create: `charts/engram-cogitator/templates/postgres-statefulset.yaml`
- Create: `charts/engram-cogitator/templates/postgres-service.yaml`
- Create: `charts/engram-cogitator/templates/postgres-secret.yaml`

**Step 1: Write Postgres StatefulSet**

```yaml
# charts/engram-cogitator/templates/postgres-statefulset.yaml
{{- if and (eq .Values.storage.driver "postgres") .Values.storage.postgres.internal }}
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ include "engram-cogitator.postgres.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
    app.kubernetes.io/component: postgres
spec:
  serviceName: {{ include "engram-cogitator.postgres.name" . }}
  replicas: 1
  selector:
    matchLabels:
      {{- include "engram-cogitator.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: postgres
  template:
    metadata:
      labels:
        {{- include "engram-cogitator.selectorLabels" . | nindent 8 }}
        app.kubernetes.io/component: postgres
    spec:
      containers:
        - name: postgres
          image: pgvector/pgvector:pg16
          env:
            - name: POSTGRES_USER
              value: {{ .Values.storage.postgres.username }}
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.storage.postgres.existingSecret | default (printf "%s-postgres" (include "engram-cogitator.fullname" .)) }}
                  key: {{ .Values.storage.postgres.existingSecretKey | default "password" }}
            - name: POSTGRES_DB
              value: {{ .Values.storage.postgres.database }}
          ports:
            - name: postgres
              containerPort: 5432
              protocol: TCP
          livenessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - {{ .Values.storage.postgres.username }}
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - {{ .Values.storage.postgres.username }}
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            {{- toYaml .Values.storage.postgres.resources | nindent 12 }}
          volumeMounts:
            - name: data
              mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes:
          - ReadWriteOnce
        {{- if .Values.storage.postgres.persistence.storageClass }}
        storageClassName: {{ .Values.storage.postgres.persistence.storageClass }}
        {{- end }}
        resources:
          requests:
            storage: {{ .Values.storage.postgres.persistence.size }}
{{- end }}
```

**Step 2: Write Postgres service**

```yaml
# charts/engram-cogitator/templates/postgres-service.yaml
{{- if and (eq .Values.storage.driver "postgres") .Values.storage.postgres.internal }}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "engram-cogitator.postgres.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
    app.kubernetes.io/component: postgres
spec:
  type: ClusterIP
  ports:
    - port: 5432
      targetPort: postgres
      protocol: TCP
      name: postgres
  selector:
    {{- include "engram-cogitator.selectorLabels" . | nindent 4 }}
    app.kubernetes.io/component: postgres
{{- end }}
```

**Step 3: Write Postgres secret**

```yaml
# charts/engram-cogitator/templates/postgres-secret.yaml
{{- if and (eq .Values.storage.driver "postgres") (not .Values.storage.postgres.existingSecret) }}
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "engram-cogitator.fullname" . }}-postgres
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
type: Opaque
data:
  password: {{ .Values.storage.postgres.password | default (randAlphaNum 24) | b64enc | quote }}
{{- end }}
```

**Step 4: Commit**

```bash
git add charts/engram-cogitator/templates/postgres-*.yaml
git commit -m "feat(helm): add optional Postgres StatefulSet"
```

---

## Task 6: Create Network Policies

**Files:**
- Create: `charts/engram-cogitator/templates/networkpolicy.yaml`

**Step 1: Write network policies**

```yaml
# charts/engram-cogitator/templates/networkpolicy.yaml
{{- if .Values.networkPolicy.enabled }}
---
# API: allow ingress from anywhere, egress to postgres/mongodb and ollama
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: {{ include "engram-cogitator.api.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
spec:
  podSelector:
    matchLabels:
      {{- include "engram-cogitator.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: api
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - ports:
        - protocol: TCP
          port: 8080
  egress:
    # DNS
    - to: []
      ports:
        - protocol: UDP
          port: 53
    # Ollama
    {{- if .Values.ollama.enabled }}
    - to:
        - podSelector:
            matchLabels:
              {{- include "engram-cogitator.selectorLabels" . | nindent 14 }}
              app.kubernetes.io/component: ollama
      ports:
        - protocol: TCP
          port: 11434
    {{- end }}
    # Postgres (internal)
    {{- if and (eq .Values.storage.driver "postgres") .Values.storage.postgres.internal }}
    - to:
        - podSelector:
            matchLabels:
              {{- include "engram-cogitator.selectorLabels" . | nindent 14 }}
              app.kubernetes.io/component: postgres
      ports:
        - protocol: TCP
          port: 5432
    {{- end }}
    # Postgres (external) or MongoDB - allow all egress on relevant ports
    {{- if or (and (eq .Values.storage.driver "postgres") (not .Values.storage.postgres.internal)) (eq .Values.storage.driver "mongodb") }}
    - to: []
      ports:
        {{- if eq .Values.storage.driver "postgres" }}
        - protocol: TCP
          port: {{ .Values.storage.postgres.port }}
        {{- else }}
        - protocol: TCP
          port: 27017
        {{- end }}
    {{- end }}
---
# Ollama: only allow from API
{{- if .Values.ollama.enabled }}
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: {{ include "engram-cogitator.ollama.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
spec:
  podSelector:
    matchLabels:
      {{- include "engram-cogitator.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: ollama
  policyTypes:
    - Ingress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              {{- include "engram-cogitator.selectorLabels" . | nindent 14 }}
              app.kubernetes.io/component: api
      ports:
        - protocol: TCP
          port: 11434
{{- end }}
---
# Postgres: only allow from API
{{- if and (eq .Values.storage.driver "postgres") .Values.storage.postgres.internal }}
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: {{ include "engram-cogitator.postgres.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
spec:
  podSelector:
    matchLabels:
      {{- include "engram-cogitator.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: postgres
  policyTypes:
    - Ingress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              {{- include "engram-cogitator.selectorLabels" . | nindent 14 }}
              app.kubernetes.io/component: api
      ports:
        - protocol: TCP
          port: 5432
{{- end }}
{{- end }}
```

**Step 2: Commit**

```bash
git add charts/engram-cogitator/templates/networkpolicy.yaml
git commit -m "feat(helm): add network policies"
```

---

## Task 7: Create Supporting Resources

**Files:**
- Create: `charts/engram-cogitator/templates/serviceaccount.yaml`
- Create: `charts/engram-cogitator/templates/pdb.yaml`
- Create: `charts/engram-cogitator/templates/hpa.yaml`
- Create: `charts/engram-cogitator/templates/ingress.yaml`

**Step 1: Write ServiceAccount**

```yaml
# charts/engram-cogitator/templates/serviceaccount.yaml
{{- if .Values.serviceAccount.create }}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "engram-cogitator.serviceAccountName" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
  {{- with .Values.serviceAccount.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
```

**Step 2: Write PodDisruptionBudget**

```yaml
# charts/engram-cogitator/templates/pdb.yaml
{{- if .Values.podDisruptionBudget.enabled }}
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "engram-cogitator.api.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
spec:
  minAvailable: {{ .Values.podDisruptionBudget.minAvailable }}
  selector:
    matchLabels:
      {{- include "engram-cogitator.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: api
{{- end }}
```

**Step 3: Write HPA**

```yaml
# charts/engram-cogitator/templates/hpa.yaml
{{- if .Values.autoscaling.enabled }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "engram-cogitator.api.name" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "engram-cogitator.api.name" . }}
  minReplicas: {{ .Values.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.autoscaling.maxReplicas }}
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ .Values.autoscaling.targetCPUUtilizationPercentage }}
{{- end }}
```

**Step 4: Write Ingress**

```yaml
# charts/engram-cogitator/templates/ingress.yaml
{{- if .Values.ingress.enabled }}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ include "engram-cogitator.fullname" . }}
  labels:
    {{- include "engram-cogitator.labels" . | nindent 4 }}
  {{- with .Values.ingress.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  {{- if .Values.ingress.className }}
  ingressClassName: {{ .Values.ingress.className }}
  {{- end }}
  {{- if .Values.ingress.tls }}
  tls:
    {{- range .Values.ingress.tls }}
    - hosts:
        {{- range .hosts }}
        - {{ . | quote }}
        {{- end }}
      secretName: {{ .secretName }}
    {{- end }}
  {{- end }}
  rules:
    {{- range .Values.ingress.hosts }}
    - host: {{ .host | quote }}
      http:
        paths:
          {{- range .paths }}
          - path: {{ .path }}
            pathType: {{ .pathType }}
            backend:
              service:
                name: {{ include "engram-cogitator.api.name" $ }}
                port:
                  name: http
          {{- end }}
    {{- end }}
{{- end }}
```

**Step 5: Commit**

```bash
git add charts/engram-cogitator/templates/serviceaccount.yaml \
        charts/engram-cogitator/templates/pdb.yaml \
        charts/engram-cogitator/templates/hpa.yaml \
        charts/engram-cogitator/templates/ingress.yaml
git commit -m "feat(helm): add serviceaccount, pdb, hpa, and ingress"
```

---

## Task 8: Validate Chart

**Step 1: Lint the chart**

Run: `helm lint charts/engram-cogitator`
Expected: No errors

**Step 2: Template with default values**

Run: `helm template test charts/engram-cogitator`
Expected: Valid YAML output

**Step 3: Template with MongoDB**

Run: `helm template test charts/engram-cogitator --set storage.driver=mongodb --set storage.mongodb.uri=mongodb://test`
Expected: Valid YAML, no postgres resources

**Step 4: Commit (if any fixes needed)**

```bash
git add -A
git commit -m "fix(helm): chart validation fixes"
```

---

## Summary

After Phase 5, you'll have:

```
charts/engram-cogitator/
  Chart.yaml
  values.yaml
  .helmignore
  templates/
    _helpers.tpl
    api-deployment.yaml
    api-service.yaml
    ollama-deployment.yaml
    ollama-service.yaml
    ollama-pvc.yaml
    postgres-statefulset.yaml
    postgres-service.yaml
    postgres-secret.yaml
    networkpolicy.yaml
    serviceaccount.yaml
    pdb.yaml
    hpa.yaml
    ingress.yaml
```

**Install examples:**

```bash
# Postgres (internal)
helm install engram charts/engram-cogitator \
  --set storage.postgres.password=mysecret

# MongoDB (external - Atlas)
helm install engram charts/engram-cogitator \
  --set storage.driver=mongodb \
  --set storage.mongodb.uri="mongodb+srv://user:pass@cluster.mongodb.net"

# With ingress
helm install engram charts/engram-cogitator \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=engram.mycompany.com
```

**Next phase:** Distribution (releases, install scripts, docs).
