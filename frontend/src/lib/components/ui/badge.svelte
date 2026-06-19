<script>
  import { cn } from "$lib/utils.js";

  /** @type {'default' | 'outline' | 'destructive' | 'success' | 'muted'} */
  export let variant = "default";

  let className = "";
  export { className as class };
</script>

<span
  data-slot="badge"
  class={cn(
    "admin-cn-badge",
    variant === "outline" && "admin-cn-badge-outline",
    variant === "destructive" && "admin-cn-badge-destructive",
    variant === "success" && "admin-cn-badge-success",
    variant === "muted" && "admin-cn-badge-muted",
    className
  )}
  {...$$restProps}
>
  <slot />
</span>
