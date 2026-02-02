// useDebounce hook
import { useEffect, useState, useRef } from 'react';

export interface UseDebounceOptions {
  immediate?: boolean;  // 新增：是否立即执行第一次回调
}

export const useDebounce = <T,>(
  value: T,
  delay: number,
  options?: UseDebounceOptions
): T => {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);
  const firstUpdate = useRef(true);

  useEffect(() => {
    // 如果启用 immediate 选项，首次变更立即生效
    if (options?.immediate && firstUpdate.current) {
      setDebouncedValue(value);
      firstUpdate.current = false;
      return;
    }

    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay, options?.immediate]);

  return debouncedValue;
};
