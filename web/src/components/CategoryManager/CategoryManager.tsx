import React, { useState, useEffect } from 'react';
import { categoryServiceClient } from '../../connect';
import { create } from '@bufbuild/protobuf';
import { 
  ListCategoriesRequestSchema, 
  CreateCategoryRequestSchema, 
  UpdateCategoryRequestSchema,
  DeleteCategoryRequestSchema 
} from '../../types/proto/api/v1/category_service_pb';
import { CategorySchema } from '../../types/proto/store/note_pb';
import type { Category } from '../../types/proto/store/note_pb';
import './CategoryManager.css';

const CategoryManager: React.FC = () => {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  
  // Form state
  const [formData, setFormData] = useState({
    nameText: '',
    slug: '',
    description: '',
    visible: true,
  });

  useEffect(() => {
    fetchCategories();
  }, []);

  const fetchCategories = async () => {
    try {
      setLoading(true);
      setError(null);
      const request = create(ListCategoriesRequestSchema, {
        includeHidden: true,
        parentId: BigInt(0),
      });
      const response = await categoryServiceClient.listCategories(request);
      setCategories(response.categories || []);
    } catch (err: any) {
      console.error('Failed to fetch categories:', err);
      setError(err.message || '获取分类列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    if (!formData.nameText.trim()) {
      alert('请输入分类名称');
      return;
    }
    if (!formData.slug.trim()) {
      alert('请输入分类标识');
      return;
    }

    try {
      const category = create(CategorySchema, {
        name: '',
        id: BigInt(0),
        nameText: formData.nameText.trim(),
        slug: formData.slug.trim(),
        description: formData.description.trim(),
        parentId: BigInt(0),
        order: 0,
        visible: formData.visible,
        createdAt: BigInt(0),
        updatedAt: BigInt(0),
      });

      const request = create(CreateCategoryRequestSchema, {
        category,
      });

      await categoryServiceClient.createCategory(request);
      
      // Reset form
      setFormData({
        nameText: '',
        slug: '',
        description: '',
        visible: true,
      });
      setIsCreating(false);
      
      // Refresh list
      fetchCategories();
    } catch (err: any) {
      console.error('Failed to create category:', err);
      alert(`创建失败: ${err.message || '未知错误'}`);
    }
  };

  const handleUpdate = async (category: Category) => {
    if (!formData.nameText.trim()) {
      alert('请输入分类名称');
      return;
    }
    if (!formData.slug.trim()) {
      alert('请输入分类标识');
      return;
    }

    try {
      const updatedCategory = create(CategorySchema, {
        name: category.name || `categories/${category.id}`,
        id: category.id,
        nameText: formData.nameText.trim(),
        slug: formData.slug.trim(),
        description: formData.description.trim(),
        parentId: category.parentId || BigInt(0),
        order: category.order || 0,
        visible: formData.visible,
        createdAt: category.createdAt || BigInt(0),
        updatedAt: BigInt(Math.floor(Date.now() / 1000)),
      });

      const request = create(UpdateCategoryRequestSchema, {
        category: updatedCategory,
      });

      await categoryServiceClient.updateCategory(request);
      
      // Reset form
      setFormData({
        nameText: '',
        slug: '',
        description: '',
        visible: true,
      });
      setEditingId(null);
      
      // Refresh list
      fetchCategories();
    } catch (err: any) {
      console.error('Failed to update category:', err);
      alert(`更新失败: ${err.message || '未知错误'}`);
    }
  };

  const handleDelete = async (category: Category) => {
    if (!confirm(`确定要删除分类 "${category.nameText}" 吗？此操作无法撤销。`)) {
      return;
    }

    try {
      const request = create(DeleteCategoryRequestSchema, {
        name: category.name || `categories/${category.id}`,
      });
      await categoryServiceClient.deleteCategory(request);
      fetchCategories();
    } catch (err: any) {
      console.error('Failed to delete category:', err);
      alert(`删除失败: ${err.message || '未知错误'}`);
    }
  };

  const startEdit = (category: Category) => {
    setEditingId(Number(category.id));
    setFormData({
      nameText: category.nameText || '',
      slug: category.slug || '',
      description: category.description || '',
      visible: category.visible !== false,
    });
    setIsCreating(false);
  };

  const cancelEdit = () => {
    setEditingId(null);
    setIsCreating(false);
    setFormData({
      nameText: '',
      slug: '',
      description: '',
      visible: true,
    });
  };

  const startCreate = () => {
    setIsCreating(true);
    setEditingId(null);
    setFormData({
      nameText: '',
      slug: '',
      description: '',
      visible: true,
    });
  };

  if (loading) {
    return <div className="category-manager-loading">加载中...</div>;
  }

  return (
    <div className="category-manager">
      <div className="category-manager-header">
        <h2>分类管理</h2>
        {!isCreating && editingId === null && (
          <button className="category-manager-btn-create" onClick={startCreate}>
            + 新建分类
          </button>
        )}
      </div>

      {error && (
        <div className="category-manager-error">{error}</div>
      )}

      {/* Create/Edit Form */}
      {(isCreating || editingId !== null) && (
        <div className="category-manager-form">
          <h3>{isCreating ? '新建分类' : '编辑分类'}</h3>
          <div className="category-form-group">
            <label>分类名称 *</label>
            <input
              type="text"
              value={formData.nameText}
              onChange={(e) => setFormData({ ...formData, nameText: e.target.value })}
              placeholder="例如：技术"
            />
          </div>
          <div className="category-form-group">
            <label>分类标识 *</label>
            <input
              type="text"
              value={formData.slug}
              onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
              placeholder="例如：technology"
            />
          </div>
          <div className="category-form-group">
            <label>描述</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="分类描述（可选）"
              rows={3}
            />
          </div>
          <div className="category-form-group">
            <label>
              <input
                type="checkbox"
                checked={formData.visible}
                onChange={(e) => setFormData({ ...formData, visible: e.target.checked })}
              />
              可见
            </label>
          </div>
          <div className="category-form-actions">
            <button
              className="category-manager-btn-primary"
              onClick={() => {
                if (isCreating) {
                  handleCreate();
                } else if (editingId !== null) {
                  const category = categories.find(c => Number(c.id) === editingId);
                  if (category) {
                    handleUpdate(category);
                  }
                }
              }}
            >
              {isCreating ? '创建' : '保存'}
            </button>
            <button
              className="category-manager-btn-cancel"
              onClick={cancelEdit}
            >
              取消
            </button>
          </div>
        </div>
      )}

      {/* Categories List */}
      <div className="category-manager-list">
        {categories.length === 0 ? (
          <div className="category-manager-empty">
            <p>还没有分类，点击"新建分类"创建第一个分类吧！</p>
          </div>
        ) : (
          <table className="category-table">
            <thead>
              <tr>
                <th>名称</th>
                <th>标识</th>
                <th>描述</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {categories.map((category) => (
                <tr key={category.id?.toString()}>
                  <td>{category.nameText || '未命名'}</td>
                  <td>{category.slug || '-'}</td>
                  <td>{category.description || '-'}</td>
                  <td>
                    <span className={`category-status ${category.visible ? 'visible' : 'hidden'}`}>
                      {category.visible ? '可见' : '隐藏'}
                    </span>
                  </td>
                  <td>
                    <div className="category-actions">
                      <button
                        className="category-action-btn-edit"
                        onClick={() => startEdit(category)}
                        disabled={editingId === Number(category.id)}
                      >
                        编辑
                      </button>
                      <button
                        className="category-action-btn-delete"
                        onClick={() => handleDelete(category)}
                      >
                        删除
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default CategoryManager;
