import * as React from "react"
import {
  Button as ChakraButton,
  type ButtonProps as ChakraButtonProps,
} from "@chakra-ui/react"

type ButtonVariant =
  | "default"
  | "destructive"
  | "outline"
  | "secondary"
  | "ghost"
  | "link"
  | "gradient"

type ButtonSize = "default" | "sm" | "lg" | "icon"

export interface ButtonProps
  extends Omit<ChakraButtonProps, "variant" | "size"> {
  variant?: ButtonVariant
  size?: ButtonSize
}

const chakraVariantMap: Record<ButtonVariant, ChakraButtonProps["variant"]> = {
  default: "solid",
  destructive: "solid",
  outline: "outline",
  secondary: "solid",
  ghost: "ghost",
  link: "link",
  gradient: "solid",
}

const chakraColorMap: Record<ButtonVariant, ChakraButtonProps["colorScheme"]> = {
  default: "teal",
  destructive: "red",
  outline: "teal",
  secondary: "gray",
  ghost: "gray",
  link: "teal",
  gradient: "teal",
}

const chakraSizeMap: Record<ButtonSize, ChakraButtonProps["size"]> = {
  default: "md",
  sm: "sm",
  lg: "lg",
  icon: "md",
}

const buttonVariants = ({
  variant = "default",
  size = "default",
}: {
  variant?: ButtonVariant
  size?: ButtonSize
}) => ({
  variant: chakraVariantMap[variant],
  colorScheme: chakraColorMap[variant],
  size: chakraSizeMap[size],
})

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = "default", size = "default", children, ...props }, ref) => {
    const mapped = buttonVariants({ variant, size })
    const gradientStyles =
      variant === "gradient"
        ? {
            bgGradient: "linear(to-r, teal.500, cyan.400)",
            color: "white",
            _hover: {
              bgGradient: "linear(to-r, teal.400, cyan.300)",
            },
          }
        : undefined

    return (
      <ChakraButton ref={ref} {...mapped} {...gradientStyles} {...props}>
        {children}
      </ChakraButton>
    )
  },
)

Button.displayName = "Button"

export { Button, buttonVariants }
