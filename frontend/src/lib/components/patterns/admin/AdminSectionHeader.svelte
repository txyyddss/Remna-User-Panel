<script>
  export let title = "";
  export let description = "";
</script>

<div class="admin-dashboard-section-head">
  <h3>{title}</h3>
  {#if description}
    <small>{description}</small>
  {/if}
</div>
