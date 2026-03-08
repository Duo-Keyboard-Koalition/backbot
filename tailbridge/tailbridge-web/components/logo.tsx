interface LogoProps {
  className?: string
  size?: number
}

export function Logo({ className, size = 40 }: LogoProps) {
  return (
    <div className={className}>
      <svg width={size} height={size} viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="50" cy="50" r="48" fill="#0A0C1B" />
        <path
          d="M50 10C27.909 10 10 27.909 10 50C10 72.091 27.909 90 50 90C72.091 90 90 72.091 90 50C90 27.909 72.091 10 50 10ZM50 20C66.569 20 80 33.431 80 50C80 66.569 66.569 80 50 80C33.431 80 20 66.569 20 50C20 33.431 33.431 20 50 20Z"
          fill="url(#paint0_linear)"
        />
        <path
          d="M65 35C57.268 35 50 42.268 50 50C50 57.732 57.268 65 65 65C72.732 65 80 57.732 80 50C80 42.268 72.732 35 65 35ZM65 45C67.761 45 70 47.239 70 50C70 52.761 67.761 55 65 55C62.239 55 60 52.761 60 50C60 47.239 62.239 45 65 45Z"
          fill="url(#paint1_linear)"
        />
        <path
          d="M35 35C27.268 35 20 42.268 20 50C20 57.732 27.268 65 35 65C42.732 65 50 57.732 50 50C50 42.268 42.732 35 35 35ZM35 45C37.761 45 40 47.239 40 50C40 52.761 37.761 55 35 55C32.239 55 30 52.761 30 50C30 47.239 32.239 45 35 45Z"
          fill="url(#paint2_linear)"
        />
        <defs>
          <linearGradient id="paint0_linear" x1="10" y1="50" x2="90" y2="50" gradientUnits="userSpaceOnUse">
            <stop stopColor="#3CDFFF" />
            <stop offset="1" stopColor="#D896FF" />
          </linearGradient>
          <linearGradient id="paint1_linear" x1="50" y1="50" x2="80" y2="50" gradientUnits="userSpaceOnUse">
            <stop stopColor="#3CDFFF" />
            <stop offset="1" stopColor="#D896FF" />
          </linearGradient>
          <linearGradient id="paint2_linear" x1="20" y1="50" x2="50" y2="50" gradientUnits="userSpaceOnUse">
            <stop stopColor="#3CDFFF" />
            <stop offset="1" stopColor="#D896FF" />
          </linearGradient>
        </defs>
      </svg>
    </div>
  )
}
