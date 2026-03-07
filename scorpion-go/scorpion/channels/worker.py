"""
AuraFlow Image Agent
Pulls jobs from image-jobs queue, generates images, uploads to MinIO.
"""
import os, json, uuid, io
import redis
from bullmq import Worker, Job  # pip install bullmq
import boto3  # MinIO is S3-compatible

REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6379")
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "localhost")
MINIO_PORT = int(os.getenv("MINIO_PORT", "9000"))
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", "auraflow")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", "changeme123")
BUCKET = "auraflow-images"

s3 = boto3.client(
    "s3",
    endpoint_url=f"http://{MINIO_ENDPOINT}:{MINIO_PORT}",
    aws_access_key_id=MINIO_ACCESS_KEY,
    aws_secret_access_key=MINIO_SECRET_KEY,
)

r = redis.from_url(REDIS_URL)

async def process_job(job: Job, job_token: str):
    data = job.data
    job_id = data["jobId"]
    prompt = data["prompt"]

    print(f"[image-agent] processing {job_id}: {prompt[:60]}")

    # TODO: Replace with actual model call
    # e.g. from diffusers import StableDiffusionPipeline
    # image = pipeline(prompt).images[0]
    image_bytes = b"<placeholder-image-bytes>"  # Replace with real generation

    key = f"{job_id}.png"
    s3.put_object(Bucket=BUCKET, Key=key, Body=image_bytes, ContentType="image/png")
    url = f"http://{MINIO_ENDPOINT}:{MINIO_PORT}/{BUCKET}/{key}"

    # Notify orchestrator
    r.publish("job:complete", json.dumps({"jobId": job_id, "storageKey": key, "url": url}))
    print(f"[image-agent] done {job_id} -> {url}")
    return {"storageKey": key, "url": url}

worker = Worker("image-jobs", process_job, {"connection": {"url": REDIS_URL}})
print("[image-agent] waiting for jobs...")
worker.run()
