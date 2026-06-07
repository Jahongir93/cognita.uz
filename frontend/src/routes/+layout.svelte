<script lang="ts">
    import { onMount } from 'svelte';
    import { navigating } from '$app/stores';
    import { authStore } from '$lib/stores/auth';
    import '../app.css';

    onMount(() => { authStore.loadUser(); });
</script>

{#if $navigating}
    <div class="page-loader" aria-hidden="true">
        <div class="loader-track">
            <div class="loader-bar"></div>
        </div>
        <div class="loader-dot">🎮</div>
    </div>
{/if}

<slot />

<style>
    .page-loader {
        position: fixed;
        top: 0; left: 0; right: 0;
        z-index: 99999;
        pointer-events: none;
    }

    /* thin gradient bar */
    .loader-track {
        height: 3px;
        background: rgba(99,102,241,.15);
        overflow: hidden;
    }
    .loader-bar {
        height: 100%;
        width: 45%;
        background: linear-gradient(90deg, #6366f1, #a78bfa, #f59e0b);
        border-radius: 0 999px 999px 0;
        animation: barSlide 1.1s cubic-bezier(.4,0,.2,1) infinite;
        box-shadow: 0 0 12px rgba(139,92,246,.6);
    }
    @keyframes barSlide {
        0%   { transform: translateX(-120%); }
        100% { transform: translateX(300%);  }
    }

    /* spinning emoji in top-right */
    .loader-dot {
        position: absolute;
        top: 10px; right: 16px;
        font-size: 1.3rem;
        line-height: 1;
        animation: dotSpin 0.8s linear infinite;
        filter: drop-shadow(0 0 6px rgba(139,92,246,.7));
    }
    @keyframes dotSpin {
        from { transform: rotate(0deg) scale(1);    }
        50%  { transform: rotate(180deg) scale(1.2); }
        to   { transform: rotate(360deg) scale(1);  }
    }
</style>
