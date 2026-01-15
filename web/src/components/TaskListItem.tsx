import React from 'react';

interface TaskListItemProps extends React.InputHTMLAttributes<HTMLInputElement> {
  checked?: boolean;
}

const TaskListItem: React.FC<TaskListItemProps> = ({ checked, ...props }) => {
  // 简单的任务列表项组件，根据 checked 属性显示正确的复选框状态
  // 注意：这里没有实现交互功能，因为当前是只读模式
  return (
    <input
      type="checkbox"
      checked={checked}
      disabled={true}
      className="task-list-checkbox"
      {...props}
    />
  );
};

export default TaskListItem;
