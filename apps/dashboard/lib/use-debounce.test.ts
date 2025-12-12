import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";

import { useDebounce } from "./use-debounce";

describe("useDebounce", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should return initial value immediately", () => {
    const { result } = renderHook(() => useDebounce("initial", 500));

    expect(result.current).toBe("initial");
  });

  it("should debounce value changes", () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      {
        initialProps: { value: "initial", delay: 500 },
      }
    );

    expect(result.current).toBe("initial");

    // Change the value
    rerender({ value: "updated", delay: 500 });

    // Value should still be the initial value before delay
    expect(result.current).toBe("initial");

    // Fast-forward time by 500ms
    act(() => {
      vi.advanceTimersByTime(500);
    });

    // Value should now be updated
    expect(result.current).toBe("updated");
  });

  it("should cancel previous timeout on rapid value changes", () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      {
        initialProps: { value: "first", delay: 500 },
      }
    );

    expect(result.current).toBe("first");

    // Change value multiple times rapidly
    rerender({ value: "second", delay: 500 });
    act(() => {
      vi.advanceTimersByTime(200);
    });

    rerender({ value: "third", delay: 500 });
    act(() => {
      vi.advanceTimersByTime(200);
    });

    rerender({ value: "fourth", delay: 500 });

    // Only 400ms have passed, value should still be initial
    expect(result.current).toBe("first");

    // Fast-forward the remaining time
    act(() => {
      vi.advanceTimersByTime(500);
    });

    // Should now be the latest value
    expect(result.current).toBe("fourth");
  });

  it("should work with different delay values", () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      {
        initialProps: { value: "initial", delay: 1000 },
      }
    );

    rerender({ value: "updated", delay: 1000 });

    // 500ms is not enough for 1000ms delay
    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(result.current).toBe("initial");

    // Complete the 1000ms delay
    act(() => {
      vi.advanceTimersByTime(500);
    });

    expect(result.current).toBe("updated");
  });

  it("should work with different types", () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      {
        initialProps: { value: 123, delay: 300 },
      }
    );

    expect(result.current).toBe(123);

    rerender({ value: 456, delay: 300 });
    act(() => {
      vi.advanceTimersByTime(300);
    });

    expect(result.current).toBe(456);
  });

  it("should work with objects", () => {
    const initialObj = { name: "John", age: 30 };
    const updatedObj = { name: "Jane", age: 25 };

    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      {
        initialProps: { value: initialObj, delay: 300 },
      }
    );

    expect(result.current).toEqual(initialObj);

    rerender({ value: updatedObj, delay: 300 });
    act(() => {
      vi.advanceTimersByTime(300);
    });

    expect(result.current).toEqual(updatedObj);
  });

  it("should cleanup timeout on unmount", () => {
    const clearTimeoutSpy = vi.spyOn(global, "clearTimeout");

    const { unmount } = renderHook(() => useDebounce("test", 500));

    unmount();

    expect(clearTimeoutSpy).toHaveBeenCalled();
  });
});
