<script>
  import { cn } from "$lib/utils.js";

  let className = "";
  export { className as class };
</script>

<div data-slot="card-content" class={cn("admin-cn-card-content", className)} {...$$restProps}>
  <slot />
</div>
