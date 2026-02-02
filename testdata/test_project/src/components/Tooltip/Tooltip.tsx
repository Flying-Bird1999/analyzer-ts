// Tooltip 组件实现
import { useState, useEffect } from 'react';

export interface TooltipProps {
  content: React.ReactNode;
  children: React.ReactNode;
  placement?: 'top' | 'bottom' | 'left' | 'right';
}

export const Tooltip: React.FC<TooltipProps> = ({ content, children, placement = 'top' }) => {
  const [visible, setVisible] = useState(false);

  return (
    <div
      className="tooltip-container"
      onMouseEnter={() => setVisible(true)}
      onMouseLeave={() => setVisible(false)}
    >
      {children}
      {visible && <div className={`tooltip tooltip-${placement}`}>{content}</div>}
    </div>
  );
};

// 简单的文本提示，不再依赖 Button
export const TextTooltip: React.FC<{ tooltip: string; text?: string }> = (props) => {
  return (
    <Tooltip content={props.tooltip}>
      <span className="tooltip-text">{props.text || "Hover me"}</span>
    </Tooltip>
  );
};
