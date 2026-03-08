import { z } from 'zod'

// Tensor dimension schema
export const DimensionSchema = z.object({
  index: z.number().int().nonnegative(),
  size: z.number().int().positive(),
  name: z.string().min(1).max(100),
  description: z.string().max(500).optional(),
})

export type DimensionInput = z.infer<typeof DimensionSchema>

// Tensor metadata schema
export const TensorMetadataSchema = z.object({
  fileName: z.string(),
  shape: z.array(z.number().int().positive()),
  dtype: z.string(),
  dimensions: z.array(DimensionSchema),
})

export type TensorMetadata = z.infer<typeof TensorMetadataSchema>

// Dataset creation schema
export const CreateDatasetSchema = z.object({
  name: z.string().min(1).max(200),
  description: z.string().max(2000).optional(),
  tags: z.array(z.string()).optional(),
  isPublic: z.boolean().default(true),
  tensors: z.array(TensorMetadataSchema),
})

export type CreateDatasetInput = z.infer<typeof CreateDatasetSchema>

// Dataset update schema
export const UpdateDatasetSchema = z.object({
  name: z.string().min(1).max(200).optional(),
  description: z.string().max(2000).optional(),
  tags: z.array(z.string()).optional(),
  isPublic: z.boolean().optional(),
})

export type UpdateDatasetInput = z.infer<typeof UpdateDatasetSchema>

// Dataset query params
export const DatasetQuerySchema = z.object({
  search: z.string().optional(),
  userId: z.string().optional(),
  tags: z.array(z.string()).optional(),
  format: z.enum(['pt', 'npy', 'npz', 'safetensors', 'h5']).optional(),
  limit: z.number().int().positive().max(100).default(20),
  offset: z.number().int().nonnegative().default(0),
})

export type DatasetQuery = z.infer<typeof DatasetQuerySchema>
