<script>
  import { cn } from "$lib/utils.js";

  let className = "";
  export { className as class };
</script>

<div data-slot="card-header" class={cn("admin-cn-card-header", className)} {...$$restProps}>
  <slot />
</div>
