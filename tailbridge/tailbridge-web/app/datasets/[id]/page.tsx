'use client'

import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { Download, Eye, TrendingUp, Calendar, User, Database, Edit, Trash2, ArrowLeft, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface Dimension {
  index: number
  size: number
  name: string
  description: string | null
}

interface Tensor {
  id: string
  fileName: string
  fileSize: number
  shape: string
  dtype: string
  dimensions: Dimension[]
}

interface Dataset {
  id: string
  name: string
  description: string | null
  fileFormat: string
  totalSize: number
  downloadCount: number
  viewCount: number
  isPublic: boolean
  tags: string | null
  createdAt: string
  user: {
    id: string
    username: string | null
    email: string
  }
  tensors: Tensor[]
  _count: {
    downloads: number
  }
}

export default function DatasetDetailPage() {
  const params = useParams()
  const router = useRouter()
  const [dataset, setDataset] = useState<Dataset | null>(null)
  const [loading, setLoading] = useState(true)
  const [downloading, setDownloading] = useState(false)
  const [error, setError] = useState('')

  const datasetId = params.id as string

  useEffect(() => {
    fetchDataset()
  }, [datasetId])

  async function fetchDataset() {
    try {
      const response = await fetch(`/api/datasets/${datasetId}`)
      if (!response.ok) {
        throw new Error('Dataset not found')
      }
      const data = await response.json()
      setDataset(data)
    } catch (err: any) {
      setError(err.message || 'Failed to load dataset')
    } finally {
      setLoading(false)
    }
  }

  async function handleDownload() {
    setDownloading(true)
    try {
      const response = await fetch(`/api/datasets/${datasetId}/download`)
      if (!response.ok) {
        throw new Error('Download failed')
      }

      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${dataset?.name.replace(/[^a-z0-9]/gi, '_')}.${dataset?.tensors.length === 1 ? dataset.fileFormat : 'zip'}`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)

      // Refresh dataset to update download count
      await fetchDataset()
    } catch (err: any) {
      setError(err.message || 'Download failed')
    } finally {
      setDownloading(false)
    }
  }

  async function handleDelete() {
    if (!confirm('Are you sure you want to delete this dataset? This action cannot be undone.')) {
      return
    }

    try {
      const response = await fetch(`/api/datasets/${datasetId}`, {
        method: 'DELETE',
      })

      if (!response.ok) {
        throw new Error('Failed to delete dataset')
      }

      router.push('/')
    } catch (err: any) {
      setError(err.message || 'Failed to delete dataset')
    }
  }

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
  }

  function formatDate(date: string): string {
    return new Date(date).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-cyber-dark via-black to-cyber-dark flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-12 h-12 mx-auto mb-4 text-cyber-blue animate-spin" />
          <p className="text-cyber-light">Loading dataset...</p>
        </div>
      </div>
    )
  }

  if (error || !dataset) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-cyber-dark via-black to-cyber-dark flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error || 'Dataset not found'}</p>
          <Link href="/">
            <Button className="bg-cyber-blue/20 border border-cyber-blue text-cyber-blue hover:bg-cyber-blue/30">
              Go Home
            </Button>
          </Link>
        </div>
      </div>
    )
  }

  const isOwner = false
  const tags = dataset.tags ? JSON.parse(dataset.tags) : []

  return (
    <div className="min-h-screen bg-gradient-to-b from-cyber-dark via-black to-cyber-dark py-12">
      <div className="container mx-auto px-4 max-w-6xl">
        {/* Back Button */}
        <Link href="/">
          <Button variant="outline" className="mb-6 border-cyber-blue/30 text-cyber-blue hover:bg-cyber-blue/10">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Datasets
          </Button>
        </Link>

        {error && (
          <div className="bg-red-500/20 border border-red-500 text-red-400 px-4 py-3 rounded mb-6">
            {error}
          </div>
        )}

        {/* Header */}
        <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-8 mb-6">
          <div className="flex items-start justify-between mb-4">
            <div className="flex-1">
              <h1 className="text-4xl font-black mb-3 bg-gradient-to-r from-cyber-blue via-cyber-pink to-cyber-purple bg-clip-text text-transparent">
                {dataset.name}
              </h1>
              {dataset.description && (
                <p className="text-cyber-light/80 text-lg">{dataset.description}</p>
              )}
            </div>
            <span className="px-3 py-1 bg-cyber-blue/20 text-cyber-blue text-sm rounded font-mono">
              .{dataset.fileFormat}
            </span>
          </div>

          {/* Tags */}
          {tags.length > 0 && (
            <div className="flex flex-wrap gap-2 mb-6">
              {tags.map((tag: string) => (
                <span key={tag} className="px-3 py-1 bg-cyber-purple/20 text-cyber-purple rounded-full text-sm">
                  {tag}
                </span>
              ))}
            </div>
          )}

          {/* Stats */}
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
            <div className="bg-black/30 border border-cyber-blue/20 rounded p-3">
              <div className="flex items-center gap-2 text-cyber-light/50 text-xs mb-1">
                <Download className="w-3 h-3" />
                Downloads
              </div>
              <div className="text-cyber-blue text-xl font-bold">{dataset.downloadCount}</div>
            </div>
            <div className="bg-black/30 border border-cyber-pink/20 rounded p-3">
              <div className="flex items-center gap-2 text-cyber-light/50 text-xs mb-1">
                <Eye className="w-3 h-3" />
                Views
              </div>
              <div className="text-cyber-pink text-xl font-bold">{dataset.viewCount}</div>
            </div>
            <div className="bg-black/30 border border-cyber-green/20 rounded p-3">
              <div className="flex items-center gap-2 text-cyber-light/50 text-xs mb-1">
                <Database className="w-3 h-3" />
                Size
              </div>
              <div className="text-cyber-green text-xl font-bold">{formatBytes(dataset.totalSize)}</div>
            </div>
            <div className="bg-black/30 border border-cyber-purple/20 rounded p-3">
              <div className="flex items-center gap-2 text-cyber-light/50 text-xs mb-1">
                <User className="w-3 h-3" />
                Author
              </div>
              <div className="text-cyber-purple text-sm font-bold truncate">
                {dataset.user.username || dataset.user.email.split('@')[0]}
              </div>
            </div>
            <div className="bg-black/30 border border-cyber-light/20 rounded p-3">
              <div className="flex items-center gap-2 text-cyber-light/50 text-xs mb-1">
                <Calendar className="w-3 h-3" />
                Created
              </div>
              <div className="text-cyber-light text-xs font-bold">{formatDate(dataset.createdAt)}</div>
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-3">
            <Button
              onClick={handleDownload}
              disabled={downloading}
              className="bg-gradient-to-r from-cyber-pink to-cyber-purple hover:from-cyber-purple hover:to-cyber-pink text-white font-bold"
            >
              {downloading ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Downloading...
                </>
              ) : (
                <>
                  <Download className="w-4 h-4 mr-2" />
                  Download Dataset
                </>
              )}
            </Button>

            {isOwner && (
              <>
                <Button
                  variant="outline"
                  className="border-cyber-blue text-cyber-blue hover:bg-cyber-blue/10"
                  onClick={() => router.push(`/datasets/${datasetId}/edit`)}
                >
                  <Edit className="w-4 h-4 mr-2" />
                  Edit
                </Button>
                <Button
                  variant="outline"
                  className="border-red-500 text-red-400 hover:bg-red-500/10"
                  onClick={handleDelete}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  Delete
                </Button>
              </>
            )}
          </div>
        </div>

        {/* Tensors */}
        <div className="space-y-6">
          <h2 className="text-2xl font-bold text-cyber-light">
            Tensors ({dataset.tensors.length})
          </h2>

          {dataset.tensors.map((tensor, idx) => {
            const shape = JSON.parse(tensor.shape)
            return (
              <div key={tensor.id} className="bg-black/50 border border-cyber-pink/30 rounded-lg p-6">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="text-xl font-bold text-cyber-light mb-1">
                      {tensor.fileName}
                    </h3>
                    <p className="text-cyber-light/50 text-sm">{formatBytes(tensor.fileSize)}</p>
                  </div>
                  <div className="text-right">
                    <div className="text-cyber-light/50 text-xs mb-1">Shape</div>
                    <div className="text-cyber-purple font-mono text-lg font-bold">
                      [{shape.join(' × ')}]
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 mb-4">
                  <div className="bg-black/30 border border-cyber-green/20 rounded p-3">
                    <div className="text-cyber-light/50 text-xs mb-1">Data Type</div>
                    <div className="text-cyber-green font-mono font-bold">{tensor.dtype}</div>
                  </div>
                  <div className="bg-black/30 border border-cyber-blue/20 rounded p-3">
                    <div className="text-cyber-light/50 text-xs mb-1">Total Elements</div>
                    <div className="text-cyber-blue font-mono font-bold">
                      {shape.reduce((a: number, b: number) => a * b, 1).toLocaleString()}
                    </div>
                  </div>
                </div>

                {/* Dimension Annotations */}
                <div>
                  <h4 className="text-lg font-bold text-cyber-light mb-3">Dimension Annotations</h4>
                  <div className="space-y-2">
                    {tensor.dimensions
                      .sort((a, b) => a.index - b.index)
                      .map((dim) => (
                        <div key={dim.index} className="bg-black/30 border border-cyber-green/30 rounded p-4">
                          <div className="grid grid-cols-12 gap-4 items-center">
                            <div className="col-span-1">
                              <div className="text-cyber-light/50 text-xs">Axis</div>
                              <div className="text-cyber-green font-mono text-lg font-bold">
                                {dim.index}
                              </div>
                            </div>
                            <div className="col-span-2">
                              <div className="text-cyber-light/50 text-xs">Size</div>
                              <div className="text-cyber-purple font-mono text-lg font-bold">
                                {dim.size.toLocaleString()}
                              </div>
                            </div>
                            <div className="col-span-3">
                              <div className="text-cyber-light/50 text-xs">Name</div>
                              <div className="text-cyber-blue font-bold">{dim.name}</div>
                            </div>
                            <div className="col-span-6">
                              <div className="text-cyber-light/50 text-xs">Description</div>
                              <div className="text-cyber-light">
                                {dim.description || <span className="italic text-cyber-light/30">No description</span>}
                              </div>
                            </div>
                          </div>
                        </div>
                      ))}
                  </div>
                </div>

                {/* Visual Representation */}
                <div className="mt-4 p-4 bg-black/30 border border-cyber-blue/20 rounded">
                  <div className="text-cyber-light/50 text-xs mb-2">Tensor Structure</div>
                  <div className="font-mono text-cyber-light">
                    <span className="text-cyber-pink">{tensor.dtype}</span>
                    <span className="text-cyber-light/50">[</span>
                    {tensor.dimensions.map((dim, i) => (
                      <span key={i}>
                        <span className="text-cyber-blue">{dim.name}</span>
                        <span className="text-cyber-light/50">:</span>
                        <span className="text-cyber-purple">{dim.size}</span>
                        {i < tensor.dimensions.length - 1 && <span className="text-cyber-light/50">, </span>}
                      </span>
                    ))}
                    <span className="text-cyber-light/50">]</span>
                  </div>
                </div>
              </div>
            )
          })}
        </div>

        {/* Usage Instructions */}
        <div className="mt-8 bg-black/50 border border-cyber-green/30 rounded-lg p-6">
          <h2 className="text-2xl font-bold text-cyber-light mb-4">Usage Instructions</h2>
          <div className="space-y-4">
            <div>
              <h3 className="text-lg font-bold text-cyber-green mb-2">PyTorch</h3>
              <pre className="bg-black/50 border border-cyber-blue/20 rounded p-4 overflow-x-auto">
                <code className="text-cyber-light text-sm">
{`import torch

# Load tensor
tensor = torch.load('${dataset.tensors[0]?.fileName}')
print(f"Shape: {tensor.shape}")
print(f"Type: {tensor.dtype}")`}
                </code>
              </pre>
            </div>

            <div>
              <h3 className="text-lg font-bold text-cyber-green mb-2">NumPy</h3>
              <pre className="bg-black/50 border border-cyber-blue/20 rounded p-4 overflow-x-auto">
                <code className="text-cyber-light text-sm">
{`import numpy as np

# Load array
array = np.load('${dataset.tensors[0]?.fileName}')
print(f"Shape: {array.shape}")
print(f"Type: {array.dtype}")`}
                </code>
              </pre>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
