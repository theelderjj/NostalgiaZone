import { useState, useCallback } from 'react';

export type ViewMode = 'split' | 'focus';

export interface UseViewOptions {
  playerCount: number;
  focusedPlayerNum?: number;
}

/**
 * Custom hook for managing emulator view modes
 * Supports split-screen (2x2 grid) and focus mode (single player fullscreen)
 */
export function useView({ playerCount, focusedPlayerNum = 1 }: UseViewOptions) {
  const [viewMode, setViewMode] = useState<ViewMode>('split');
  const [focusedPlayer, setFocusedPlayer] = useState<number>(focusedPlayerNum);

  // Toggle between split and focus mode
  const toggleViewMode = useCallback(() => {
    setViewMode((prev) => (prev === 'split' ? 'focus' : 'split'));
  }, []);

  // Set which player to focus on
  const focusPlayer = useCallback((playerNum: number) => {
    setFocusedPlayer(playerNum);
    setViewMode('focus');
  }, []);

  // Get CSS class for the current view mode
  const getContainerClass = useCallback(() => {
    return viewMode === 'split' ? 'split-2x2' : 'split-focus';
  }, [viewMode]);

  // Check if a specific player should be visible in current view mode
  const isPlayerVisible = useCallback(
    (playerNum: number) => {
      if (viewMode === 'split') {
        return true; // All players visible in split mode
      }
      return playerNum === focusedPlayer; // Only focused player in focus mode
    },
    [viewMode, focusedPlayer]
  );

  // Get grid position for a player in split mode
  const getPlayerGridPosition = useCallback(
    (playerNum: number) => {
      if (viewMode === 'focus') {
        return { row: 1, col: 1 };
      }

      // 2x2 grid layout
      const positions = [
        { row: 1, col: 1 }, // Player 1 - top left
        { row: 1, col: 2 }, // Player 2 - top right
        { row: 2, col: 1 }, // Player 3 - bottom left
        { row: 2, col: 2 }, // Player 4 - bottom right
      ];

      return positions[playerNum - 1] || { row: 1, col: 1 };
    },
    [viewMode]
  );

  // Reset to split view
  const resetView = useCallback(() => {
    setViewMode('split');
  }, []);

  return {
    viewMode,
    focusedPlayer,
    toggleViewMode,
    focusPlayer,
    resetView,
    getContainerClass,
    isPlayerVisible,
    getPlayerGridPosition,
  };
}
