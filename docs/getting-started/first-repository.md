# Creating Your First Repository

## Step 1: Start GitForge

Ensure the web server and at least one Git backend (HTTP or SSH) are running.

## Step 2: Create a Repository

Navigate to the web interface and create a new repository, or use the API:

```bash
curl -X POST http://localhost:3000/api/repos \
  -H "Content-Type: application/json" \
  -d '{"name": "my-repo", "description": "My first repo"}'
```

## Step 3: Push Your Code

```bash
cd your-project
git init
git remote add origin http://localhost:3000/username/my-repo.git
git add .
git commit -m "Initial commit"
git push -u origin main
```

## Step 4: Clone an Existing Repository

```bash
git clone http://localhost:3000/username/my-repo.git
```
