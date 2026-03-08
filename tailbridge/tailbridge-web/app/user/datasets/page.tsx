'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { Database, Download, TrendingUp, Edit, Trash2, Plus, Loader2, Eye } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface Dataset {
  id: string
  name: string
  description: string
  fileFormat: string
  totalSize: number
  downloadCount: number
  viewCount: number
  isPublic: boolean
  createdAt: string
  tensors: Array<{
    shape: string
    dtype: string
  }>
}

export default function MyDatasetsPage() {
  const [datasets, setDatasets] = useState<Dataset[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchMyDatasets()
  }, [])

  async function fetchMyDatasets() {
    try {
      const response = await fetch('/api/datasets')
      const data = await response.json()
      setDatasets(data.datasets || [])
    } catch (error) {
      console.error('Error fetching datasets:', error)
    } finally {
      setLoading(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Are you sure you want to delete this dataset? This action cannot be undone.')) {
      return
    }

    try {
      const response = await fetch(`/api/datasets/${id}`, {
        method: 'DELETE',
      })

      if (!response.ok) {
        throw new Error('Failed to delete dataset')
      }

      setDatasets(datasets.filter(d => d.id !== id))
    } catch (error) {
      console.error('Error deleting dataset:', error)
      alert('Failed to delete dataset')
    }
  }

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="w-8 h-8 animate-spin text-cyber-blue" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-cyber-dark via-black to-cyber-dark py-12 px-4">
      <div className="container mx-auto max-w-7xl">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-4xl font-black mb-2 bg-gradient-to-r from-cyber-blue via-cyber-pink to-cyber-purple bg-clip-text text-transparent">
              My Datasets
            </h1>
            <p className="text-cyber-light/70">
              Manage your tensor datasets
            </p>
          </div>

          <Link href="/upload">
            <Button className="bg-gradient-to-r from-cyber-pink to-cyber-purple hover:from-cyber-purple hover:to-cyber-pink text-white font-bold">
              <Plus className="w-4 h-4 mr-2" />
              New Dataset
            </Button>
          </Link>
        </div>

        {datasets.length === 0 ? (
          <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-12 text-center">
            <Database className="w-16 h-16 mx-auto mb-4 text-cyber-light/30" />
            <h3 className="text-xl font-bold text-cyber-light mb-2">No datasets yet</h3>
            <p className="text-cyber-light/70 mb-6">
              Upload your first tensor dataset to get started
            </p>
            <Link href="/upload">
              <Button className="bg-gradient-to-r from-cyber-pink to-cyber-purple hover:from-cyber-purple hover:to-cyber-pink text-white font-bold">
                <Plus className="w-4 h-4 mr-2" />
                Upload Dataset
              </Button>
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {datasets.map((dataset) => (
              <div
                key={dataset.id}
                className="bg-black/50 border border-cyber-blue/30 rounded-lg p-6 hover:border-cyber-pink transition-all hover:shadow-[0_0_20px_rgba(255,0,128,0.3)]"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1">
                    <h3 className="text-xl font-bold text-cyber-light mb-1">{dataset.name}</h3>
                    <div className="flex items-center gap-2">
                      <span className="px-2 py-1 bg-cyber-blue/20 text-cyber-blue text-xs rounded font-mono">
                        .{dataset.fileFormat}
                      </span>
                      <span className={`px-2 py-1 text-xs rounded ${dataset.isPublic ? 'bg-cyber-green/20 text-cyber-green' : 'bg-cyber-light/20 text-cyber-light'}`}>
                        {dataset.isPublic ? 'Public' : 'Private'}
                      </span>
                    </div>
                  </div>
                </div>

                <p className="text-cyber-light/70 text-sm mb-4 line-clamp-2">
                  {dataset.description || 'No description provided'}
                </p>

                {dataset.tensors.length > 0 && (
                  <div className="mb-4 text-xs">
                    <div className="text-cyber-light/50">Shape:</div>
                    <div className="text-cyber-purple font-mono">
                      {JSON.parse(dataset.tensors[0].shape).join(' × ')}
                    </div>
                    <div className="text-cyber-light/50 mt-1">
                      Type: <span className="text-cyber-green">{dataset.tensors[0].dtype}</span>
                    </div>
                  </div>
                )}

                <div className="flex items-center justify-between text-xs text-cyber-light/50 mb-4">
                  <span>{formatBytes(dataset.totalSize)}</span>
                  <span>{new Date(dataset.createdAt).toLocaleDateString()}</span>
                </div>

                <div className="flex items-center gap-4 mb-4 text-xs text-cyber-light/50">
                  <span className="flex items-center gap-1">
                    <Download className="w-3 h-3" />
                    {dataset.downloadCount}
                  </span>
                  <span className="flex items-center gap-1">
                    <Eye className="w-3 h-3" />
                    {dataset.viewCount}
                  </span>
                </div>

                <div className="flex gap-2">
                  <Link href={`/datasets/${dataset.id}`} className="flex-1">
                    <Button
                      variant="outline"
                      className="w-full border-cyber-blue text-cyber-blue hover:bg-cyber-blue/10"
                    >
                      <Eye className="w-4 h-4 mr-2" />
                      View
                    </Button>
                  </Link>
                  <Button
                    variant="outline"
                    className="border-red-500 text-red-400 hover:bg-red-500/10"
                    onClick={() => handleDelete(dataset.id)}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
