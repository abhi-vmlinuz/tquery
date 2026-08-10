# Practical Examples & DevOps Pipelines

`tquery` integrates directly into Unix pipelines, cloud CLI tools, and development workflows.

---

## 1. LLM & Inference API Endpoints

### Query Available Models
```bash
curl -s https://inference.dahl.global/v1/models | tquery
```

### Select Specific Model Fields
```bash
curl -s https://inference.dahl.global/v1/models | \
  tquery '.data[] | {id, owned_by, created}'
```

---

## 2. GitHub REST API

### List Recent Releases in a Clean Table
```bash
curl -s https://api.github.com/repos/charmbracelet/bubbletea/releases | \
  tquery '.[0:5] | .[] | {tag: .tag_name, name: .name, published: .published_at, prerelease}'
```

### Inspect Open Pull Requests
```bash
gh pr list --json number,title,author,headRefName | tquery
```

---

## 3. Docker & Container Management

### Inspect Docker Container Configurations
```bash
docker inspect my-container | tquery '.[0] | {Id: .Id[0:12], Image: .Config.Image, Running: .State.Running}'
```

### Inspect Docker Network IP Allocations
```bash
docker network inspect bridge | tquery '.[0].Containers | to_entries[] | {id: .key[0:12], name: .value.Name, ip: .value.IPv4Address}'
```

---

## 4. Kubernetes (`kubectl`)

### Visualize Pod Resource Limits as a Markdown Table
```bash
kubectl get pods -o json | \
  tquery -f markdown '.items[] | {pod: .metadata.name, namespace: .metadata.namespace, status: .status.phase}'
```

---

## 5. AWS CLI

### List EC2 Instances
```bash
aws ec2 describe-instances --output json | \
  tquery '.Reservations[].Instances[] | {InstanceId, InstanceType, State: .State.Name, PublicIp: .PublicIpAddress}'
```

### List S3 Buckets
```bash
aws s3api list-buckets --output json | tquery '.Buckets'
```

---

## 6. Local File Workflows

### Convert Any Complex JSON File to CSV
```bash
tquery -f csv '.records' analytics_dump.json > report.csv
```

### Display Nested Configuration as a Tree
```bash
tquery -f tree package.json
```

### Interactively Explore Deep JSON Dumps
```bash
tquery -i large_api_response.json
```
