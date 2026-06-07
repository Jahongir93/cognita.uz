<script lang="ts">
  import CategoryPageView from '$lib/components/CategoryPageView.svelte';
  import { allCategories } from '$lib/data/categories';
  const category = allCategories.find(c => c.id === 'attestation')!;
</script>

<CategoryPageView {category} />
