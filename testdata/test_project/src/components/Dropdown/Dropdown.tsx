// Dropdown 组件实现
import { useState } from 'react';

export interface DropdownProps {
  trigger: React.ReactNode;
  items: string[];
  onSelect?: (item: string) => void;
}

export const Dropdown: React.FC<DropdownProps> = ({ trigger, items, onSelect }) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="dropdown">
      <div onClick={() => setIsOpen(!isOpen)}>{trigger}</div>
      {isOpen && (
        <ul className="dropdown-menu">
          {items.map((item, index) => (
            <li key={index} onClick={() => { onSelect?.(item); setIsOpen(false); }}>
              {item}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

// 简单的下拉触发器，不再依赖 Button
export const SimpleDropdown: React.FC<{ items: string[]; label?: string }> = (props) => {
  return (
    <Dropdown
      trigger={<span className="dropdown-trigger">{props.label || "Select"}</span>}
      items={props.items}
    />
  );
};
