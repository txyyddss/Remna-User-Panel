<script>
  import { cn } from "$lib/utils.js";

  let className = "";
  export { className as class };
</script>

<div data-slot="card-title" class={cn("admin-cn-card-title", className)} {...$$restProps}>
  <slot />
</div>
