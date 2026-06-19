<script>
  import { cva } from "class-variance-authority";
  import { cn } from "$lib/utils.js";

  export let type = "button";
  export let variant = "default";
  export let size = "default";
  export let disabled = false;
  export let href = "";
  export let onclick = undefined;
  let className = "";
  export { className as class };

  const buttonVariants = cva("btn", {
    variants: {
      variant: {
        default: "btn-primary",
        secondary: "btn-secondary",
        outline: "btn-outline",
        ghost: "btn-ghost",
        telegram: "btn-telegram",
        icon: "btn-icon",
      },
      size: {
        default: "",
        sm: "btn-sm",
        lg: "btn-lg",
        icon: "btn-square",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  });
</script>

{#if href}
  <a class={cn(buttonVariants({ variant, size }), className)} {href} {onclick} {...$$restProps}>
    <slot />
  </a>
{:else}
  <button
    class={cn(buttonVariants({ variant, size }), className)}
    {type}
    {disabled}
    {onclick}
    {...$$restProps}
  >
    <slot />
  </button>
{/if}
