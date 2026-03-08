'use client'

import { ChakraProvider, extendTheme } from '@chakra-ui/react'
import type { ReactNode } from 'react'

const theme = extendTheme({
  fonts: {
    heading: 'Orbitron, sans-serif',
    body: 'Inter, sans-serif',
  },
})

export function ChakraAppProvider({ children }: { children: ReactNode }) {
  return <ChakraProvider theme={theme}>{children}</ChakraProvider>
}
