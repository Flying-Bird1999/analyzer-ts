// Tabs 组件实现
import { useState } from 'react';

export interface TabItem {
  label: string;
  content: React.ReactNode;
}

export interface TabsProps {
  items: TabItem[];
  defaultIndex?: number;
}

export const Tabs: React.FC<TabsProps> = ({ items, defaultIndex = 0 }) => {
  const [activeIndex, setActiveIndex] = useState(defaultIndex);

  return (
    <div className="tabs">
      <div className="tab-headers">
        {items.map((item, index) => (
          <button
            key={index}
            className={`tab-header ${index === activeIndex ? 'active' : ''}`}
            onClick={() => setActiveIndex(index)}
          >
            {item.label}
          </button>
        ))}
      </div>
      <div className="tab-content">
        {items[activeIndex]?.content}
      </div>
    </div>
  );
};

// 简单的标签切换，不再依赖 Button
export const SimpleTabs: React.FC<{ tabs: string[] }> = (props) => {
  const [activeTab, setActiveTab] = useState(0);

  return (
    <div className="simple-tabs">
      {props.tabs.map((tab, index) => (
        <span
          key={index}
          className={`simple-tab ${activeTab === index ? 'active' : ''}`}
          onClick={() => setActiveTab(index)}
        >
          {tab}
        </span>
      ))}
    </div>
  );
};
