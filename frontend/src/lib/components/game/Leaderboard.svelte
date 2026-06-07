<script lang="ts">
    import type { LeaderboardEntry } from '$lib/api/types';

    export let entries: LeaderboardEntry[];
    export let myId: string = '';
    export let title: string = 'Natijalar';
    export let showTop: number = 10;

    $: visible = entries.slice(0, showTop);
    $: myEntry = entries.find(e => e.id === myId);
    $: myIsInTop = myEntry ? myEntry.rank <= showTop : false;

    function medalEmoji(rank: number): string {
        if (rank === 1) return '🥇';
        if (rank === 2) return '🥈';
        if (rank === 3) return '🥉';
        return `#${rank}`;
    }
</script>

<div class="leaderboard">
    <h2>{title}</h2>

    <ul class="list">
        {#each visible as entry (entry.id)}
            <li class="entry" class:me={entry.id === myId} class:top3={entry.rank <= 3}>
                <span class="rank">{medalEmoji(entry.rank)}</span>
                <span class="avatar">{entry.avatar}</span>
                <span class="name">{entry.nickname}</span>
                <span class="streak" title="Streak" aria-label="{entry.streak} streak">
                    {#if entry.streak >= 2}🔥×{entry.streak}{/if}
                </span>
                <div class="score-col">
                    <span class="score">{entry.score.toLocaleString()}</span>
                    {#if entry.delta !== 0}
                        <span class="delta" class:up={entry.delta > 0} class:down={entry.delta < 0}>
                            {entry.delta > 0 ? `▲${entry.delta}` : `▼${Math.abs(entry.delta)}`}
                        </span>
                    {/if}
                </div>
            </li>
        {/each}
    </ul>

    <!-- Show own rank if outside top N -->
    {#if myEntry && !myIsInTop}
        <div class="separator">···</div>
        <ul class="list">
            <li class="entry me">
                <span class="rank">#{myEntry.rank}</span>
                <span class="avatar">{myEntry.avatar}</span>
                <span class="name">{myEntry.nickname} (Siz)</span>
                <span class="streak"></span>
                <div class="score-col">
                    <span class="score">{myEntry.score.toLocaleString()}</span>
                </div>
            </li>
        </ul>
    {/if}
</div>

<style>
    .leaderboard {
        width: 100%;
        max-width: 520px;
        margin: 0 auto;
    }
    h2 {
        text-align: center;
        font-size: 1.5rem;
        font-weight: 700;
        margin-bottom: 16px;
        color: #1e293b;
    }
    .list {
        list-style: none;
        padding: 0;
        margin: 0;
        display: flex;
        flex-direction: column;
        gap: 6px;
    }
    .entry {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 10px 14px;
        background: #f8fafc;
        border-radius: 10px;
        border: 2px solid transparent;
        transition: transform 0.3s;
    }
    .entry.me {
        background: #eff6ff;
        border-color: #3b82f6;
        font-weight: 700;
    }
    .entry.top3 {
        background: #fefce8;
    }
    .rank {
        width: 40px;
        text-align: center;
        font-size: 1.1rem;
        font-weight: 700;
        flex-shrink: 0;
    }
    .avatar {
        font-size: 1.5rem;
        flex-shrink: 0;
    }
    .name {
        flex: 1;
        font-size: 0.95rem;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    .streak {
        font-size: 0.8rem;
        color: #f59e0b;
        flex-shrink: 0;
        min-width: 40px;
    }
    .score-col {
        display: flex;
        flex-direction: column;
        align-items: flex-end;
        flex-shrink: 0;
    }
    .score {
        font-size: 1rem;
        font-weight: 700;
        color: #1e293b;
    }
    .delta {
        font-size: 0.7rem;
    }
    .delta.up { color: #22c55e; }
    .delta.down { color: #ef4444; }
    .separator {
        text-align: center;
        color: #94a3b8;
        margin: 4px 0;
    }
</style>
