<script lang="ts">
    export let seconds: number;
    export let maxSeconds: number;
    export let size: number = 80;

    $: percent = maxSeconds > 0 ? (seconds / maxSeconds) * 100 : 0;
    $: strokeColor = percent > 50 ? '#22c55e' : percent > 25 ? '#f59e0b' : '#ef4444';
    $: isUrgent = seconds <= 5;

    const radius = 30;
    const circumference = 2 * Math.PI * radius;
    $: offset = circumference - (percent / 100) * circumference;
</script>

<div
    class="timer"
    class:urgent={isUrgent}
    style="width:{size}px;height:{size}px"
    aria-label="{seconds} seconds left"
>
    <svg width={size} height={size} viewBox="0 0 80 80">
        <!-- Background circle -->
        <circle cx="40" cy="40" r={radius} fill="none" stroke="#e5e7eb" stroke-width="6" />
        <!-- Progress circle -->
        <circle
            cx="40" cy="40" r={radius}
            fill="none"
            stroke={strokeColor}
            stroke-width="6"
            stroke-linecap="round"
            stroke-dasharray={circumference}
            stroke-dashoffset={offset}
            transform="rotate(-90 40 40)"
            style="transition: stroke-dashoffset 0.9s linear, stroke 0.3s"
        />
    </svg>
    <span class="label" style="color:{strokeColor}">{seconds}</span>
</div>

<style>
    .timer {
        position: relative;
        display: flex;
        align-items: center;
        justify-content: center;
    }
    svg {
        position: absolute;
        top: 0;
        left: 0;
    }
    .label {
        font-size: 1.4rem;
        font-weight: 700;
        position: relative;
        z-index: 1;
    }
    .urgent .label {
        animation: pulse 0.5s infinite alternate;
    }
    @keyframes pulse {
        from { transform: scale(1); }
        to { transform: scale(1.2); }
    }
</style>
