from typing import Any, Dict, Optional
import asyncio
from ..core import ToolParameter, ToolParameterType

async def upload_document_handler(agent: Any, file_path: str) -> str:
    """Upload a document to the assistant and wait for indexing"""
    if not agent.client or not agent.assistant_id:
        return "Error: Assistant not initialized"
    
    try:
        # Upload document
        document = await agent.client.upload_document_to_assistant(
            agent.assistant_id,
            file_path
        )
        
        doc_id = document.document_id
        # Wait for indexing (simple polling for the tool)
        # Note: In a real agent, we might want to return the ID and let the agent poll
        # or just return that it's started. But here we'll wait for a bit to make it "interactive".
        
        max_retries = 30
        for _ in range(max_retries):
            status = await agent.client.get_document_status(doc_id)
            if status.status == "indexed":
                return f"Document '{file_path}' uploaded and indexed successfully. ID: {doc_id}"
            elif status.status == "failed":
                return f"Document indexing failed: {status.status_message}"
            await asyncio.sleep(2)
            
        return f"Document upload started, but indexing is taking longer than expected. ID: {doc_id}"
    except Exception as e:
        return f"Error uploading document: {str(e)}"

async def get_document_status_handler(agent: Any, document_id: str) -> str:
    """Check the status of a document"""
    if not agent.client:
        return "Error: Client not initialized"
        
    try:
        status = await agent.client.get_document_status(document_id)
        return f"Status for {document_id}: {status.status}. Message: {status.status_message or 'N/A'}"
    except Exception as e:
        return f"Error getting document status: {str(e)}"

def get_document_tools(agent: Any):
    return [
        {
            "name": "upload_document",
            "description": "Upload a document (PDF, TXT, etc.) to your current assistant context for analysis.",
            "parameters": [
                ToolParameter("file_path", ToolParameterType.STRING, "Path to the file to upload")
            ],
            "handler": lambda file_path: upload_document_handler(agent, file_path)
        },
        {
            "name": "get_document_status",
            "description": "Check if a previously uploaded document has finished indexing.",
            "parameters": [
                ToolParameter("document_id", ToolParameterType.STRING, "The ID of the document to check")
            ],
            "handler": lambda document_id: get_document_status_handler(agent, document_id)
        }
    ]
