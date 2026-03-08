'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { Search, Upload, Database, TrendingUp, Download } from 'lucide-react'
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
  _count: {
    downloads: number
  }
}

export default function HomePage() {
  const [datasets, setDatasets] = useState<Dataset[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')

  useEffect(() => {
    fetchDatasets()
  }, [])

  async function fetchDatasets(search?: string) {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      if (search) params.append('search', search)
      params.append('limit', '12')

      const response = await fetch(`/api/datasets?${params.toString()}`)
      const data = await response.json()
      setDatasets(data.datasets || [])
    } catch (error) {
      console.error('Error fetching datasets:', error)
    } finally {
      setLoading(false)
    }
  }

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    fetchDatasets(searchQuery)
  }

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-cyber-dark via-black to-cyber-dark">
      {/* Hero Section */}
      <div className="relative overflow-hidden">
        <div className="absolute inset-0 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:14px_24px]"></div>

        <div className="container mx-auto px-4 py-16 relative">
          <div className="text-center max-w-4xl mx-auto mb-12">
            <h1 className="text-6xl font-black mb-6 bg-gradient-to-r from-cyber-blue via-cyber-pink to-cyber-purple bg-clip-text text-transparent">
              TensorStore
            </h1>
            <p className="text-lg text-cyber-light/60 mb-2">Kaggle for Tensors</p>
            <p className="text-2xl font-bold text-cyber-pink mb-4">
              The Next Evolution of Programming
            </p>
            <p className="text-xl text-cyber-light/80 mb-4">
              Upload and share pure tensor datasets. No data engineering needed - just vectors ready for training.
            </p>
            <p className="text-base text-cyber-light/70 max-w-3xl mx-auto mb-8">
              As brain-inspired AI models, embeddings, and vector representations dominate modern computing, we're witnessing the shift from
              <span className="text-cyber-blue font-semibold"> Object-Oriented Programming</span> to
              <span className="text-cyber-purple font-semibold"> Tensor-Oriented Programming</span>.
              TensorStore is the foundation of this new paradigm - where everything is vectors, and data is already ML-ready.
            </p>

            <div className="flex gap-4 justify-center">
              <Link href="/upload">
                <Button className="bg-gradient-to-r from-cyber-pink to-cyber-purple hover:from-cyber-purple hover:to-cyber-pink text-white font-bold py-6 px-8 text-lg hover:animate-pulse">
                  <Upload className="mr-2" />
                  Upload Dataset
                </Button>
              </Link>
              <Link href="/datasets">
                <Button variant="outline" className="border-cyber-blue text-cyber-blue hover:bg-cyber-blue/10 font-bold py-6 px-8 text-lg">
                  <Database className="mr-2" />
                  Browse Datasets
                </Button>
              </Link>
              <Link href="/docs">
                <Button variant="outline" className="border-cyber-pink text-cyber-pink hover:bg-cyber-pink/10 font-bold py-6 px-8 text-lg">
                  Setup Docs
                </Button>
              </Link>
            </div>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl mx-auto mb-16">
            <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-6 text-center backdrop-blur-sm">
              <Database className="w-12 h-12 mx-auto mb-3 text-cyber-blue" />
              <div className="text-3xl font-bold text-cyber-blue">{datasets.length}</div>
              <div className="text-cyber-light/60">Datasets</div>
            </div>
            <div className="bg-black/50 border border-cyber-pink/30 rounded-lg p-6 text-center backdrop-blur-sm">
              <Download className="w-12 h-12 mx-auto mb-3 text-cyber-pink" />
              <div className="text-3xl font-bold text-cyber-pink">
                {datasets.reduce((acc, d) => acc + d.downloadCount, 0)}
              </div>
              <div className="text-cyber-light/60">Downloads</div>
            </div>
            <div className="bg-black/50 border border-cyber-purple/30 rounded-lg p-6 text-center backdrop-blur-sm">
              <TrendingUp className="w-12 h-12 mx-auto mb-3 text-cyber-purple" />
              <div className="text-3xl font-bold text-cyber-purple">100%</div>
              <div className="text-cyber-light/60">Vector Ready</div>
            </div>
          </div>

          {/* Search Bar */}
          <form onSubmit={handleSearch} className="max-w-2xl mx-auto mb-12">
            <div className="relative">
              <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 text-cyber-light/50" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search datasets by name or description..."
                className="w-full pl-12 pr-4 py-4 bg-black/50 border border-cyber-blue/30 rounded-lg text-cyber-light placeholder:text-cyber-light/30 focus:border-cyber-blue focus:outline-none"
              />
            </div>
          </form>

          {/* Recent Datasets */}
          <div className="max-w-7xl mx-auto">
            <h2 className="text-3xl font-bold text-cyber-light mb-6">Recent Datasets</h2>

            {loading ? (
              <div className="text-center py-12 text-cyber-light/60">Loading datasets...</div>
            ) : datasets.length === 0 ? (
              <div className="text-center py-12 text-cyber-light/60">
                No datasets found. Be the first to upload one!
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {datasets.map((dataset) => (
                  <Link key={dataset.id} href={`/datasets/${dataset.id}`}>
                    <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-6 hover:border-cyber-pink transition-all hover:shadow-[0_0_20px_rgba(255,0,128,0.3)] cursor-pointer h-full">
                      <div className="flex items-start justify-between mb-3">
                        <h3 className="text-xl font-bold text-cyber-light">{dataset.name}</h3>
                        <span className="px-2 py-1 bg-cyber-blue/20 text-cyber-blue text-xs rounded">
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
                          <div className="text-cyber-light/50 mt-1">Type: <span className="text-cyber-green">{dataset.tensors[0].dtype}</span></div>
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
            )}
          </div>
        </div>
      </div>

      {/* Paradigm Shift Section */}
      <div className="border-t border-cyber-blue/20 py-16">
        <div className="container mx-auto px-4">
          <div className="max-w-4xl mx-auto text-center mb-16">
            <h2 className="text-4xl font-bold text-cyber-light mb-6">
              The Tensor-Oriented Programming Paradigm
            </h2>
            <p className="text-lg text-cyber-light/80 mb-8">
              With the explosive growth of brain-inspired AI models, neural embeddings, and vector databases, we're experiencing a fundamental shift in how we represent and process information.
            </p>
            <div className="bg-black/50 border border-cyber-purple/30 rounded-lg p-8">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
                <div className="text-left">
                  <h3 className="text-xl font-bold text-cyber-light/50 mb-4 line-through">Object-Oriented Era</h3>
                  <ul className="space-y-2 text-cyber-light/60">
                    <li>• Classes and objects</li>
                    <li>• String-based data</li>
                    <li>• Manual feature engineering</li>
                    <li>• CSV files and databases</li>
                  </ul>
                </div>
                <div className="text-left">
                  <h3 className="text-xl font-bold text-cyber-pink mb-4">Tensor-Oriented Future</h3>
                  <ul className="space-y-2 text-cyber-light">
                    <li>• <span className="text-cyber-blue">Vectors and embeddings</span></li>
                    <li>• <span className="text-cyber-purple">Everything encoded</span></li>
                    <li>• <span className="text-cyber-green">Ready for neural networks</span></li>
                    <li>• <span className="text-cyber-pink">Pure tensor datasets</span></li>
                  </ul>
                </div>
              </div>
              <p className="text-cyber-light/70 text-sm">
                Every piece of data - text, images, audio, video - is now represented as vectors.
                Brain-inspired AI models have made tensor representation the default. TensorStore embraces this reality.
              </p>
            </div>
          </div>

          <h2 className="text-3xl font-bold text-center text-cyber-light mb-12">Why TensorStore?</h2>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-5xl mx-auto">
            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-gradient-to-br from-cyber-blue to-cyber-purple flex items-center justify-center">
                <Database className="w-8 h-8 text-white" />
              </div>
              <h3 className="text-xl font-bold text-cyber-light mb-2">Pure Tensors Only</h3>
              <p className="text-cyber-light/70">
                No messy CSV files or unstructured data. Everything is already in tensor format, ready for PyTorch or NumPy.
                This is how data should exist in the age of brain-inspired AI.
              </p>
            </div>

            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-gradient-to-br from-cyber-pink to-cyber-purple flex items-center justify-center">
                <Search className="w-8 h-8 text-white" />
              </div>
              <h3 className="text-xl font-bold text-cyber-light mb-2">Semantic Dimensions</h3>
              <p className="text-cyber-light/70">
                Each dimension is annotated with its meaning - batch size, channels, height, width, embedding dimensions.
                Context is preserved in the tensor metadata.
              </p>
            </div>

            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-gradient-to-br from-cyber-purple to-cyber-green flex items-center justify-center">
                <Upload className="w-8 h-8 text-white" />
              </div>
              <h3 className="text-xl font-bold text-cyber-light mb-2">Vector-First World</h3>
              <p className="text-cyber-light/70">
                Upload .pt, .npy, .safetensors, or .h5 files. Your embeddings, encodings, and representations
                are first-class citizens, not afterthoughts.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
