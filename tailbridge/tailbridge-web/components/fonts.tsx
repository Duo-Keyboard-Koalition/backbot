"use client"

import { Global } from "@emotion/react"

export function Fonts() {
  return (
    <Global
      styles={`
        @font-face {
          font-family: 'Minecraft';
          src: url('/fonts/minecraft.woff2') format('woff2');
          font-weight: normal;
          font-style: normal;
          font-display: swap;
        }
      `}
    />
  )
}
