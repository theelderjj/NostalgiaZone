import { describe, it, expect, vi } from 'vitest';
import { useView } from '../src/hooks/useView';
import { renderHook, act } from '@testing-library/react';

describe('useView', () => {
  it('starts in split mode', () => {
    const { result } = renderHook(() => useView({ playerCount: 4 }));
    
    expect(result.current.viewMode).toBe('split');
    expect(result.current.focusedPlayer).toBe(1);
  });

  it('toggles between split and focus mode', () => {
    const { result } = renderHook(() => useView({ playerCount: 4 }));
    
    expect(result.current.viewMode).toBe('split');
    
    act(() => {
      result.current.toggleViewMode();
    });
    
    expect(result.current.viewMode).toBe('focus');
    
    act(() => {
      result.current.toggleViewMode();
    });
    
    expect(result.current.viewMode).toBe('split');
  });

  it('focuses on specific player', () => {
    const { result } = renderHook(() => useView({ playerCount: 4 }));
    
    act(() => {
      result.current.focusPlayer(3);
    });
    
    expect(result.current.viewMode).toBe('focus');
    expect(result.current.focusedPlayer).toBe(3);
  });

  it('shows all players in split mode', () => {
    const { result } = renderHook(() => useView({ playerCount: 4 }));
    
    expect(result.current.isPlayerVisible(1)).toBe(true);
    expect(result.current.isPlayerVisible(2)).toBe(true);
    expect(result.current.isPlayerVisible(3)).toBe(true);
    expect(result.current.isPlayerVisible(4)).toBe(true);
  });

  it('shows only focused player in focus mode', () => {
    const { result } = renderHook(() => useView({ playerCount: 4 }));
    
    act(() => {
      result.current.focusPlayer(2);
    });
    
    expect(result.current.isPlayerVisible(1)).toBe(false);
    expect(result.current.isPlayerVisible(2)).toBe(true);
    expect(result.current.isPlayerVisible(3)).toBe(false);
    expect(result.current.isPlayerVisible(4)).toBe(false);
  });

  it('returns correct grid positions for split mode', () => {
    const { result } = renderHook(() => useView({ playerCount: 4 }));
    
    expect(result.current.getPlayerGridPosition(1)).toEqual({ row: 1, col: 1 });
    expect(result.current.getPlayerGridPosition(2)).toEqual({ row: 1, col: 2 });
    expect(result.current.getPlayerGridPosition(3)).toEqual({ row: 2, col: 1 });
    expect(result.current.getPlayerGridPosition(4)).toEqual({ row: 2, col: 2 });
  });

  it('resets to split view', () => {
    const { result } = renderHook(() => useView({ playerCount: 4 }));
    
    act(() => {
      result.current.focusPlayer(1);
    });
    
    expect(result.current.viewMode).toBe('focus');
    
    act(() => {
      result.current.resetView();
    });
    
    expect(result.current.viewMode).toBe('split');
  });

  it('returns correct container class', () => {
    const { result } = renderHook(() => useView({ playerCount: 4 }));
    
    expect(result.current.getContainerClass()).toBe('split-2x2');
    
    act(() => {
      result.current.toggleViewMode();
    });
    
    expect(result.current.getContainerClass()).toBe('split-focus');
  });
});
