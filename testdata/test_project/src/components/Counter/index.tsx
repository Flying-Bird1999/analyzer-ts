// Counter 组件 - 引用 hooks 的 default export 用于测试
import useCounter from '../../hooks/useCounter';

export const Counter = () => {
  const { count, increment, decrement, reset } = useCounter(0);

  return (
    <div>
      <span>Count: {count}</span>
      <button onClick={increment}>+</button>
      <button onClick={decrement}>-</button>
      <button onClick={reset}>Reset</button>
    </div>
  );
};
