// useCounter - 用于测试 default export 场景
import { useState } from 'react';

interface UseCounterResult {
  count: number;
  increment: () => void;
  decrement: () => void;
  reset: () => void;
}

// 测试 export default function foo() {} 语法
export default function useCounter(initialValue: number = 0): UseCounterResult {
  const [count, setCount] = useState(initialValue);

  const increment = () => setCount(prev => prev + 1);
  const decrement = () => setCount(prev => prev - 1);
  const reset = () => setCount(initialValue);

  return { count, increment, decrement, reset };
}
