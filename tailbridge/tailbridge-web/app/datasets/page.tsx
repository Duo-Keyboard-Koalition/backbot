'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { Search, Filter, Download, TrendingUp, Database } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface Dataset {
  id: string
  name: string
  description: string
  fileFormat: string
  totalSize: number
  downloadCount: number
  viewCount: number
  createdAt: string
  user: {
    username: string
    email: string
  }
  tensors: Array<{
    shape: string
    dtype: string
  }>
}

const FORMATS = ['pt', 'npy', 'npz', 'safetensors', 'h5']

export default function DatasetsPage() {
  const [datasets, setDatasets] = useState<Dataset[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedFormat, setSelectedFormat] = useState<string>('')
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const limit = 12

  useEffect(() => {
    fetchDatasets()
  }, [searchQuery, selectedFormat, offset])

  async function fetchDatasets() {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      if (searchQuery) params.append('search', searchQuery)
      if (selectedFormat) params.append('format', selectedFormat)
      params.append('limit', limit.toString())
      params.append('offset', offset.toString())

      const response = await fetch(`/api/datasets?${params.toString()}`)
      const data = await response.json()
      setDatasets(data.datasets || [])
      setTotal(data.total || 0)
    } catch (error) {
      console.error('Error fetching datasets:', error)
    } finally {
      setLoading(false)
    }
  }

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    setOffset(0)
    fetchDatasets()
  }

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
  }

  const totalPages = Math.ceil(total / limit)
  const currentPage = Math.floor(offset / limit) + 1

  return (
    <div className="min-h-screen bg-gradient-to-b from-cyber-dark via-black to-cyber-dark py-12">
      <div className="container mx-auto px-4 max-w-7xl">
        <div className="mb-8">
          <h1 className="text-4xl font-black mb-4 bg-gradient-to-r from-cyber-blue via-cyber-pink to-cyber-purple bg-clip-text text-transparent">
            Browse Datasets
          </h1>
          <p className="text-cyber-light/70">
            Explore {total} tensor datasets ready for machine learning
          </p>
        </div>

        {/* Search and Filter */}
        <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-6 mb-8">
          <form onSubmit={handleSearch} className="space-y-4">
            <div className="flex gap-4">
              <div className="flex-1 relative">
                <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 text-cyber-light/50" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search datasets..."
                  className="w-full pl-12 pr-4 py-3 bg-black/50 border border-cyber-blue/30 rounded text-cyber-light placeholder:text-cyber-light/30 focus:border-cyber-blue focus:outline-none"
                />
              </div>
              <Button
                type="submit"
                className="bg-cyber-blue/20 border border-cyber-blue text-cyber-blue hover:bg-cyber-blue/30"
              >
                Search
              </Button>
            </div>

            <div className="flex items-center gap-4">
              <Filter className="text-cyber-light/50" />
              <div className="flex gap-2 flex-wrap">
                <button
                  type="button"
                  onClick={() => setSelectedFormat('')}
                  className={`px-4 py-2 rounded transition-all ${
                    selectedFormat === ''
                      ? 'bg-cyber-blue/30 border border-cyber-blue text-cyber-blue'
                      : 'bg-black/30 border border-cyber-light/20 text-cyber-light/50 hover:border-cyber-light/40'
                  }`}
                >
                  All Formats
                </button>
                {FORMATS.map((format) => (
                  <button
                    key={format}
                    type="button"
                    onClick={() => setSelectedFormat(format)}
                    className={`px-4 py-2 rounded transition-all font-mono ${
                      selectedFormat === format
                        ? 'bg-cyber-blue/30 border border-cyber-blue text-cyber-blue'
                        : 'bg-black/30 border border-cyber-light/20 text-cyber-light/50 hover:border-cyber-light/40'
                    }`}
                  >
                    .{format}
                  </button>
                ))}
              </div>
            </div>
          </form>
        </div>

        {/* Results */}
        {loading ? (
          <div className="text-center py-12 text-cyber-light/60">Loading datasets...</div>
        ) : datasets.length === 0 ? (
          <div className="text-center py-12">
            <Database className="w-16 h-16 mx-auto mb-4 text-cyber-light/30" />
            <p className="text-cyber-light/60 mb-4">No datasets found</p>
            <Link href="/upload">
              <Button className="bg-gradient-to-r from-cyber-pink to-cyber-purple hover:from-cyber-purple hover:to-cyber-pink text-white font-bold">
                Upload First Dataset
              </Button>
            </Link>
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
              {datasets.map((dataset) => (
                <Link key={dataset.id} href={`/datasets/${dataset.id}`}>
                  <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-6 hover:border-cyber-pink transition-all hover:shadow-[0_0_20px_rgba(255,0,128,0.3)] cursor-pointer h-full">
                    <div className="flex items-start justify-between mb-3">
                      <h3 className="text-xl font-bold text-cyber-light">{dataset.name}</h3>
                      <span className="px-2 py-1 bg-cyber-blue/20 text-cyber-blue text-xs rounded font-mono">
                        .{dataset.fileFormat}
                      </span>
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

                    <div className="flex items-center justify-between text-xs text-cyber-light/50">
                      <span>By {dataset.user.username || dataset.user.email.split('@')[0]}</span>
                      <span>{formatBytes(dataset.totalSize)}</span>
                    </div>

                    <div className="flex items-center gap-4 mt-3 text-xs text-cyber-light/50">
                      <span className="flex items-center gap-1">
                        <Download className="w-3 h-3" />
                        {dataset.downloadCount}
                      </span>
                      <span className="flex items-center gap-1">
                        <TrendingUp className="w-3 h-3" />
                        {dataset.viewCount}
                      </span>
                    </div>
                  </div>
                </Link>
              ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center gap-4">
                <Button
                  onClick={() => setOffset(Math.max(0, offset - limit))}
                  disabled={offset === 0}
                  variant="outline"
                  className="border-cyber-blue text-cyber-blue hover:bg-cyber-blue/10 disabled:opacity-30"
                >
                  Previous
                </Button>
                <span className="text-cyber-light">
                  Page {currentPage} of {totalPages}
                </span>
                <Button
                  onClick={() => setOffset(offset + limit)}
                  disabled={offset + limit >= total}
                  variant="outline"
                  className="border-cyber-blue text-cyber-blue hover:bg-cyber-blue/10 disabled:opacity-30"
                >
                  Next
                </Button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
