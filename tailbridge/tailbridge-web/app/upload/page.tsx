'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Upload, X, Plus, Trash2, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface UploadedFile {
  file: File
  fileName: string
  originalName: string
}

interface DimensionInput {
  index: number
  size: number
  name: string
  description: string
}

interface TensorMetadata {
  fileName: string
  shape: number[]
  dtype: string
  dimensions: DimensionInput[]
}

const DTYPES = [
  'float16', 'float32', 'float64',
  'int8', 'int16', 'int32', 'int64',
  'uint8', 'uint16', 'uint32', 'uint64',
  'bool', 'complex64', 'complex128'
]

export default function UploadPage() {
  const router = useRouter()

  const [uploadedFiles, setUploadedFiles] = useState<UploadedFile[]>([])
  const [datasetName, setDatasetName] = useState('')
  const [datasetDescription, setDatasetDescription] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [tagInput, setTagInput] = useState('')
  const [isPublic, setIsPublic] = useState(true)

  const [tensorMetadata, setTensorMetadata] = useState<TensorMetadata[]>([])
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || [])
    if (files.length === 0) return

    setError('')
    setUploading(true)

    try {
      const formData = new FormData()
      files.forEach(file => formData.append('files', file))

      const response = await fetch('/api/upload', {
        method: 'POST',
        body: formData,
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Upload failed')
      }

      const data = await response.json()
      const newFiles: UploadedFile[] = data.files.map((f: any) => ({
        file: files.find(file => file.name === f.originalName)!,
        fileName: f.fileName,
        originalName: f.originalName,
      }))

      setUploadedFiles([...uploadedFiles, ...newFiles])

      // Initialize metadata for new files
      const newMetadata: TensorMetadata[] = newFiles.map(f => ({
        fileName: f.fileName,
        shape: [],
        dtype: 'float32',
        dimensions: [],
      }))
      setTensorMetadata([...tensorMetadata, ...newMetadata])
    } catch (err: any) {
      setError(err.message || 'Failed to upload files')
    } finally {
      setUploading(false)
    }
  }

  const removeFile = (index: number) => {
    setUploadedFiles(uploadedFiles.filter((_, i) => i !== index))
    setTensorMetadata(tensorMetadata.filter((_, i) => i !== index))
  }

  const updateShape = (fileIndex: number, shapeStr: string) => {
    const shape = shapeStr.split(',').map(s => parseInt(s.trim())).filter(n => !isNaN(n))
    const updated = [...tensorMetadata]
    updated[fileIndex].shape = shape

    // Auto-generate dimensions based on shape
    if (shape.length > 0) {
      const commonNames = ['batch', 'channels', 'height', 'width', 'depth', 'features']
      updated[fileIndex].dimensions = shape.map((size, idx) => ({
        index: idx,
        size,
        name: commonNames[idx] || `dim_${idx}`,
        description: '',
      }))
    }

    setTensorMetadata(updated)
  }

  const updateDtype = (fileIndex: number, dtype: string) => {
    const updated = [...tensorMetadata]
    updated[fileIndex].dtype = dtype
    setTensorMetadata(updated)
  }

  const updateDimension = (fileIndex: number, dimIndex: number, field: 'name' | 'description', value: string) => {
    const updated = [...tensorMetadata]
    updated[fileIndex].dimensions[dimIndex][field] = value
    setTensorMetadata(updated)
  }

  const addTag = () => {
    if (tagInput.trim() && !tags.includes(tagInput.trim())) {
      setTags([...tags, tagInput.trim()])
      setTagInput('')
    }
  }

  const removeTag = (tag: string) => {
    setTags(tags.filter(t => t !== tag))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // Validation
    if (!datasetName.trim()) {
      setError('Dataset name is required')
      return
    }

    if (uploadedFiles.length === 0) {
      setError('At least one tensor file is required')
      return
    }

    // Validate all tensors have shape and dimensions
    for (let i = 0; i < tensorMetadata.length; i++) {
      const tensor = tensorMetadata[i]
      if (tensor.shape.length === 0) {
        setError(`Tensor ${i + 1} (${uploadedFiles[i].originalName}) needs a shape`)
        return
      }
      if (tensor.dimensions.length === 0) {
        setError(`Tensor ${i + 1} (${uploadedFiles[i].originalName}) needs dimension annotations`)
        return
      }
      for (let j = 0; j < tensor.dimensions.length; j++) {
        if (!tensor.dimensions[j].name.trim()) {
          setError(`Tensor ${i + 1}, dimension ${j} needs a name`)
          return
        }
      }
    }

    setUploading(true)

    try {
      const response = await fetch('/api/datasets', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: datasetName,
          description: datasetDescription,
          tags,
          isPublic,
          tensors: tensorMetadata,
        }),
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Failed to create dataset')
      }

      const dataset = await response.json()
      router.push(`/datasets/${dataset.id}`)
    } catch (err: any) {
      setError(err.message || 'Failed to create dataset')
      setUploading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-cyber-dark via-black to-cyber-dark py-12">
      <div className="container mx-auto px-4 max-w-4xl">
        <h1 className="text-4xl font-black mb-8 bg-gradient-to-r from-cyber-blue via-cyber-pink to-cyber-purple bg-clip-text text-transparent">
          Upload Tensor Dataset
        </h1>

        {error && (
          <div className="bg-red-500/20 border border-red-500 text-red-400 px-4 py-3 rounded mb-6">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-8">
          {/* Dataset Metadata */}
          <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-6">
            <h2 className="text-2xl font-bold text-cyber-light mb-4">Dataset Information</h2>

            <div className="space-y-4">
              <div>
                <label className="block text-cyber-light mb-2">Dataset Name *</label>
                <input
                  type="text"
                  value={datasetName}
                  onChange={(e) => setDatasetName(e.target.value)}
                  className="w-full px-4 py-2 bg-black/50 border border-cyber-blue/30 rounded text-cyber-light focus:border-cyber-blue focus:outline-none"
                  placeholder="e.g., MNIST Training Set"
                  required
                />
              </div>

              <div>
                <label className="block text-cyber-light mb-2">Description</label>
                <textarea
                  value={datasetDescription}
                  onChange={(e) => setDatasetDescription(e.target.value)}
                  rows={4}
                  className="w-full px-4 py-2 bg-black/50 border border-cyber-blue/30 rounded text-cyber-light focus:border-cyber-blue focus:outline-none"
                  placeholder="Describe your dataset..."
                />
              </div>

              <div>
                <label className="block text-cyber-light mb-2">Tags</label>
                <div className="flex gap-2 mb-2">
                  <input
                    type="text"
                    value={tagInput}
                    onChange={(e) => setTagInput(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), addTag())}
                    className="flex-1 px-4 py-2 bg-black/50 border border-cyber-blue/30 rounded text-cyber-light focus:border-cyber-blue focus:outline-none"
                    placeholder="Add tags (press Enter)"
                  />
                  <Button type="button" onClick={addTag} className="bg-cyber-blue/20 border border-cyber-blue text-cyber-blue hover:bg-cyber-blue/30">
                    <Plus className="w-4 h-4" />
                  </Button>
                </div>
                <div className="flex flex-wrap gap-2">
                  {tags.map(tag => (
                    <span key={tag} className="px-3 py-1 bg-cyber-purple/20 text-cyber-purple rounded-full text-sm flex items-center gap-2">
                      {tag}
                      <button type="button" onClick={() => removeTag(tag)}>
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  ))}
                </div>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="isPublic"
                  checked={isPublic}
                  onChange={(e) => setIsPublic(e.target.checked)}
                  className="w-4 h-4"
                />
                <label htmlFor="isPublic" className="text-cyber-light">Make this dataset public</label>
              </div>
            </div>
          </div>

          {/* File Upload */}
          <div className="bg-black/50 border border-cyber-blue/30 rounded-lg p-6">
            <h2 className="text-2xl font-bold text-cyber-light mb-4">Upload Tensor Files</h2>

            <div className="border-2 border-dashed border-cyber-blue/30 rounded-lg p-8 text-center mb-4">
              <Upload className="w-12 h-12 mx-auto mb-4 text-cyber-blue" />
              <label className="cursor-pointer">
                <span className="text-cyber-light">
                  Drop files here or <span className="text-cyber-blue underline">browse</span>
                </span>
                <input
                  type="file"
                  multiple
                  accept=".pt,.pth,.npy,.npz,.safetensors,.h5,.hdf5"
                  onChange={handleFileUpload}
                  className="hidden"
                  disabled={uploading}
                />
              </label>
              <p className="text-cyber-light/50 text-sm mt-2">
                Supported: .pt, .npy, .npz, .safetensors, .h5
              </p>
            </div>

            {uploading && uploadedFiles.length === 0 && (
              <div className="flex items-center justify-center gap-2 text-cyber-blue">
                <Loader2 className="w-5 h-5 animate-spin" />
                Uploading files...
              </div>
            )}

            {uploadedFiles.map((file, fileIndex) => (
              <div key={fileIndex} className="bg-black/30 border border-cyber-pink/30 rounded-lg p-6 mb-4">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="text-lg font-bold text-cyber-light">{file.originalName}</h3>
                    <p className="text-sm text-cyber-light/50">{(file.file.size / 1024 / 1024).toFixed(2)} MB</p>
                  </div>
                  <button
                    type="button"
                    onClick={() => removeFile(fileIndex)}
                    className="text-red-400 hover:text-red-300"
                  >
                    <Trash2 className="w-5 h-5" />
                  </button>
                </div>

                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-cyber-light mb-2">Shape (comma-separated) *</label>
                      <input
                        type="text"
                        value={tensorMetadata[fileIndex]?.shape.join(', ') || ''}
                        onChange={(e) => updateShape(fileIndex, e.target.value)}
                        className="w-full px-4 py-2 bg-black/50 border border-cyber-blue/30 rounded text-cyber-light focus:border-cyber-blue focus:outline-none font-mono"
                        placeholder="e.g., 128, 256, 3"
                        required
                      />
                      <p className="text-xs text-cyber-light/50 mt-1">
                        Example: 60000, 28, 28 for [batch, height, width]
                      </p>
                    </div>

                    <div>
                      <label className="block text-cyber-light mb-2">Data Type *</label>
                      <select
                        value={tensorMetadata[fileIndex]?.dtype || 'float32'}
                        onChange={(e) => updateDtype(fileIndex, e.target.value)}
                        className="w-full px-4 py-2 bg-black/50 border border-cyber-blue/30 rounded text-cyber-light focus:border-cyber-blue focus:outline-none"
                        required
                      >
                        {DTYPES.map(dtype => (
                          <option key={dtype} value={dtype}>{dtype}</option>
                        ))}
                      </select>
                    </div>
                  </div>

                  {/* Dimension Annotations */}
                  {tensorMetadata[fileIndex]?.dimensions.length > 0 && (
                    <div>
                      <h4 className="text-lg font-bold text-cyber-light mb-3">Dimension Annotations</h4>
                      <div className="space-y-3">
                        {tensorMetadata[fileIndex].dimensions.map((dim, dimIndex) => (
                          <div key={dimIndex} className="bg-black/30 border border-cyber-green/30 rounded p-4">
                            <div className="grid grid-cols-12 gap-3 items-start">
                              <div className="col-span-1">
                                <label className="block text-cyber-light/50 text-xs mb-1">Axis</label>
                                <div className="text-cyber-green font-mono text-lg font-bold">{dimIndex}</div>
                              </div>
                              <div className="col-span-1">
                                <label className="block text-cyber-light/50 text-xs mb-1">Size</label>
                                <div className="text-cyber-purple font-mono text-lg font-bold">{dim.size}</div>
                              </div>
                              <div className="col-span-4">
                                <label className="block text-cyber-light/50 text-xs mb-1">Name *</label>
                                <input
                                  type="text"
                                  value={dim.name}
                                  onChange={(e) => updateDimension(fileIndex, dimIndex, 'name', e.target.value)}
                                  className="w-full px-3 py-1 bg-black/50 border border-cyber-blue/30 rounded text-cyber-light text-sm focus:border-cyber-blue focus:outline-none"
                                  placeholder="e.g., batch"
                                  required
                                />
                              </div>
                              <div className="col-span-6">
                                <label className="block text-cyber-light/50 text-xs mb-1">Description</label>
                                <input
                                  type="text"
                                  value={dim.description}
                                  onChange={(e) => updateDimension(fileIndex, dimIndex, 'description', e.target.value)}
                                  className="w-full px-3 py-1 bg-black/50 border border-cyber-blue/30 rounded text-cyber-light text-sm focus:border-cyber-blue focus:outline-none"
                                  placeholder="e.g., Number of samples in batch"
                                />
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>

          {/* Submit Button */}
          <div className="flex justify-end gap-4">
            <Button
              type="button"
              onClick={() => router.push('/')}
              variant="outline"
              className="border-cyber-light/30 text-cyber-light hover:bg-cyber-light/10"
              disabled={uploading}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              className="bg-gradient-to-r from-cyber-pink to-cyber-purple hover:from-cyber-purple hover:to-cyber-pink text-white font-bold"
              disabled={uploading || uploadedFiles.length === 0}
            >
              {uploading ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Creating Dataset...
                </>
              ) : (
                <>
                  <Upload className="w-4 h-4 mr-2" />
                  Create Dataset
                </>
              )}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
