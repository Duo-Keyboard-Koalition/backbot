# Image Agent

Pulls from the `image-jobs` queue, generates images, uploads to MinIO.

## Models Supported
- Stable Diffusion (via `diffusers`)
- ComfyUI API
- DALL-E 3 (OpenAI, not open source but drop-in)

## Config (env vars)
| Var | Default | Description |
|-----|---------|-------------|
| REDIS_URL | redis://localhost:6379 | Redis connection |
| DATABASE_URL | - | PostgreSQL |
| MINIO_ENDPOINT | localhost | MinIO host |
| MINIO_ACCESS_KEY | - | MinIO key |
| MINIO_SECRET_KEY | - | MinIO secret |
| MODEL_BACKEND | comfyui | comfyui \| diffusers \| openai |
| COMFYUI_URL | http://localhost:8188 | ComfyUI API endpoint |
| CONCURRENCY | 2 | Parallel jobs |
