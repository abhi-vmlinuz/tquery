# Practical Examples & DevOps Pipelines

`tq` integrates directly into Unix pipelines, cloud CLI tools, and development workflows.

---

## 1. LLM & Inference API Endpoints

### Query Available Models
```bash
curl -s https://integrate.api.nvidia.com/v1/models | tq
```

### Limit to First 10 Models
```bash
curl -s https://integrate.api.nvidia.com/v1/models | tq -10
```

### Search Specific Models with Regex
```bash
curl -s https://integrate.api.nvidia.com/v1/models | tq -g 'deepseek|llama|meta'
```

### Select Specific Model Fields (without manual JQ projection)
```bash
curl -s https://integrate.api.nvidia.com/v1/models | tq -c id,owned_by,created
```

### Sort Models by Timestamp (Ascending / Descending)
```bash
# Newest models first (descending by timestamp)
curl -s https://integrate.api.nvidia.com/v1/models | tq -s -created -10

# Or with explicit --desc
curl -s https://integrate.api.nvidia.com/v1/models | tq -c id,created -s created --desc -10
```

---

## 2. GitHub REST API

### List Recent Releases in a Clean Table
```bash
curl -s https://api.github.com/repos/charmbracelet/bubbletea/releases | \
  tq -5 '.[0:5] | .[] | {tag: .tag_name, name: .name, published: .published_at, prerelease}'
```

### Inspect Open Pull Requests
```bash
gh pr list --json number,title,author,headRefName | tq
```

---

## 3. Docker & Container Management

### Inspect Docker Container Configurations
```bash
docker inspect my-container | tq '.[0] | {Id: .Id[0:12], Image: .Config.Image, Running: .State.Running}'
```

### Inspect Docker Network IP Allocations
```bash
docker network inspect bridge | tq '.[0].Containers | to_entries[] | {id: .key[0:12], name: .value.Name, ip: .value.IPv4Address}'
```

---

## 4. Kubernetes (`kubectl`)

### Visualize Pod Resource Limits as a Markdown Table
```bash
kubectl get pods -o json | \
  tq -f markdown '.items[] | {pod: .metadata.name, namespace: .metadata.namespace, status: .status.phase}'
```

---

## 5. AWS CLI

### List EC2 Instances
```bash
aws ec2 describe-instances --output json | \
  tq '.Reservations[].Instances[] | {InstanceId, InstanceType, State: .State.Name, PublicIp: .PublicIpAddress}'
```

### List S3 Buckets
```bash
aws s3api list-buckets --output json | tq '.Buckets'
```

---

## 6. Local File Workflows

### Convert Any Complex JSON File to CSV
```bash
tq -f csv '.records' analytics_dump.json > report.csv
```

### Display Nested Configuration as a Tree
```bash
tq -f tree package.json
```

### Interactively Explore Deep JSON Dumps
```bash
tq -i large_api_response.json
```

---

## 7. Clipboard Workflows

### Inspect Copied API Responses Instantly
```bash
# View JSON in clipboard without saving to a temp file
tq -cb

# Filter errors in clipboard JSON
tq -cb -g 'error'

# Project specific columns from clipboard JSON
tq -cb -c id,status
```

