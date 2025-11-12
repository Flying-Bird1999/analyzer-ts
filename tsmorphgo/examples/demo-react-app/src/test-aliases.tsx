/**
 * 测试别名映射的文件
 * 使用 @/ 别名导入组件和工具
 */

import { formatDate } from '@/utils/dateUtils';

// 模拟组件
const TestComponent: React.FC = () => {
  const today = new Date();

  return (
    <div>
      <h1>别名映射测试</h1>
      <p>格式化日期: {formatDate(today)}</p>
    </div>
  );
};

export default TestComponent;