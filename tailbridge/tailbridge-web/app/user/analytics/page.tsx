'use client'

import { useEffect, useState } from 'react'
import { TrendingUp, Download, Eye, Database, Calendar, Loader2 } from 'lucide-react'

interface Analytics {
  datasets: Array<{
    id: string
    name: string
    downloadCount: number
    viewCount: number
    createdAt: string
  }>
  totalDownloads: number
  totalViews: number
  avgDownloadsPerDataset: number
  avgViewsPerDataset: number
}

export default function AnalyticsPage() {
  const [analytics, setAnalytics] = useState<Analytics | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchAnalytics()
  }, [])

  async function fetchAnalytics() {
    try {
      const response = await fetch('/api/datasets')
      const data = await response.json()

      const datasets = data.datasets || []
      const totalDownloads = datasets.reduce((acc: number, d: any) => acc + d.downloadCount, 0)
      const totalViews = datasets.reduce((acc: number, d: any) => acc + d.viewCount, 0)

      setAnalytics({
        datasets,
        totalDownloads,
        totalViews,
        avgDownloadsPerDataset: datasets.length > 0 ? totalDownloads / datasets.length : 0,
        avgViewsPerDataset: datasets.length > 0 ? totalViews / datasets.length : 0,
      })
    } catch (error) {
      console.error('Error fetching analytics:', error)
    } finally {
      setLoading(false)
    }
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
        <h1 className="text-4xl font-black mb-8 bg-gradient-to-r from-cyber-blue via-cyber-pink to-cyber-purple bg-clip-text text-transparent">
          Analytics
        </h1>

        {/* Overview Stats */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-6">
            <div className="flex items-center justify-between mb-2">
              <Database className="w-8 h-8 text-cyber-blue" />
            </div>
            <div className="text-3xl font-bold text-cyber-blue">{analytics?.datasets.length || 0}</div>
            <div className="text-sm text-cyber-light/70">Total Datasets</div>
          </div>

          <div className="bg-black/50 border border-cyber-pink/30 rounded-lg p-6">
            <div className="flex items-center justify-between mb-2">
              <Download className="w-8 h-8 text-cyber-pink" />
            </div>
            <div className="text-3xl font-bold text-cyber-pink">{analytics?.totalDownloads || 0}</div>
            <div className="text-sm text-cyber-light/70">Total Downloads</div>
          </div>

          <div className="bg-black/50 border border-cyber-purple/30 rounded-lg p-6">
            <div className="flex items-center justify-between mb-2">
              <Eye className="w-8 h-8 text-cyber-purple" />
            </div>
            <div className="text-3xl font-bold text-cyber-purple">{analytics?.totalViews || 0}</div>
            <div className="text-sm text-cyber-light/70">Total Views</div>
          </div>

          <div className="bg-black/50 border border-cyber-green/30 rounded-lg p-6">
            <div className="flex items-center justify-between mb-2">
              <TrendingUp className="w-8 h-8 text-cyber-green" />
            </div>
            <div className="text-3xl font-bold text-cyber-green">
              {Math.round(analytics?.avgDownloadsPerDataset || 0)}
            </div>
            <div className="text-sm text-cyber-light/70">Avg Downloads</div>
          </div>
        </div>

        {/* Dataset Performance */}
        <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-8">
          <h2 className="text-2xl font-bold text-cyber-light mb-6">Dataset Performance</h2>

          {!analytics?.datasets.length ? (
            <div className="text-center py-12 text-cyber-light/50">
              <TrendingUp className="w-16 h-16 mx-auto mb-4 opacity-30" />
              <p>No datasets to analyze</p>
              <p className="text-sm mt-2">Upload datasets to see their performance metrics</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-cyber-blue/30">
                    <th className="text-left py-3 px-4 text-cyber-light/70 font-medium">Dataset</th>
                    <th className="text-right py-3 px-4 text-cyber-light/70 font-medium">Downloads</th>
                    <th className="text-right py-3 px-4 text-cyber-light/70 font-medium">Views</th>
                    <th className="text-right py-3 px-4 text-cyber-light/70 font-medium">Engagement</th>
                    <th className="text-right py-3 px-4 text-cyber-light/70 font-medium">Created</th>
                  </tr>
                </thead>
                <tbody>
                  {analytics.datasets
                    .sort((a, b) => b.downloadCount - a.downloadCount)
                    .map((dataset) => {
                      const engagement = dataset.viewCount > 0
                        ? ((dataset.downloadCount / dataset.viewCount) * 100).toFixed(1)
                        : '0.0'

                      return (
                        <tr key={dataset.id} className="border-b border-cyber-blue/10 hover:bg-cyber-blue/5">
                          <td className="py-3 px-4 text-cyber-light">{dataset.name}</td>
                          <td className="py-3 px-4 text-right">
                            <span className="text-cyber-pink font-bold">{dataset.downloadCount}</span>
                          </td>
                          <td className="py-3 px-4 text-right">
                            <span className="text-cyber-purple font-bold">{dataset.viewCount}</span>
                          </td>
                          <td className="py-3 px-4 text-right">
                            <span className="text-cyber-green font-bold">{engagement}%</span>
                          </td>
                          <td className="py-3 px-4 text-right text-cyber-light/50 text-sm">
                            {new Date(dataset.createdAt).toLocaleDateString()}
                          </td>
                        </tr>
                      )
                    })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
