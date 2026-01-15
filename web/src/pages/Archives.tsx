import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

interface NoteItem {
  id: number;
  title: string;
  slug: string;
  createdAt: string;
}

interface ArchiveGroup {
  year: string;
  notes: NoteItem[];
}

const Archives: React.FC = () => {
  const [archiveData, setArchiveData] = useState<ArchiveGroup[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Fetch notes from API and group by year
    const fetchArchives = async () => {
      try {
        // Replace with actual API call
        const mockNotes: NoteItem[] = [
          {
            id: 1,
            title: "Go CLI 开发利器：Cobra 简明教程",
            slug: "go-cli-cobra-tutorial",
            createdAt: "2025/08/30"
          },
          {
            id: 2,
            title: "Go实战指南：使用 go-redis 执行 Lua 脚本",
            slug: "go-redis-lua-script",
            createdAt: "2025/07/15"
          },
          {
            id: 3,
            title: "基于泛型的轻量级依赖注入工具 do",
            slug: "golang-generic-di-tool-do",
            createdAt: "2025/06/30"
          },
          {
            id: 4,
            title: "使用 gzip 拯救你的 varchar",
            slug: "use-gzip-save-varchar",
            createdAt: "2025/06/04"
          },
          {
            id: 5,
            title: "使用 chromedp 操作 chrome",
            slug: "use-chromedp-operate-chrome",
            createdAt: "2025/04/06"
          },
          {
            id: 6,
            title: "pulsar 介绍及Pulsar Go client 使用指南",
            slug: "pulsar-go-client-guide",
            createdAt: "2025/03/31"
          },
          {
            id: 7,
            title: "[译]Go Protobuf：新的 Opaque API",
            slug: "go-protobuf-opaque-api",
            createdAt: "2025/01/31"
          },
          {
            id: 8,
            title: "Go语言中的迭代器和 iter 包",
            slug: "go-iterator-iter-package",
            createdAt: "2024/12/24"
          },
          {
            id: 9,
            title: "SQL优先的 Go ORM 框架——Bun 介绍",
            slug: "sql-first-golang-orm-bun",
            createdAt: "2024/11/30"
          },
          {
            id: 10,
            title: "ORM 框架 ent 介绍",
            slug: "golang-orm-ent-introduction",
            createdAt: "2024/11/07"
          },
          {
            id: 11,
            title: "[译] Prometheus 运算符",
            slug: "prometheus-operator",
            createdAt: "2024/08/04"
          },
          {
            id: 12,
            title: "[译]查询 Prometheus",
            slug: "query-prometheus",
            createdAt: "2024/07/16"
          },
          {
            id: 13,
            title: "Prometheus 介绍",
            slug: "prometheus-introduction",
            createdAt: "2024/07/15"
          },
          {
            id: 14,
            title: "GORM配置链路追踪",
            slug: "gorm-tracing-configuration",
            createdAt: "2024/04/14"
          },
          {
            id: 15,
            title: "go-redis配置链路追踪",
            slug: "go-redis-tracing-configuration",
            createdAt: "2024/04/14"
          },
          {
            id: 16,
            title: "zap日志库配置链路追踪",
            slug: "zap-log-tracing-configuration",
            createdAt: "2024/04/14"
          },
          {
            id: 17,
            title: "gRPC的链路追踪",
            slug: "grpc-tracing",
            createdAt: "2024/04/08"
          },
          {
            id: 18,
            title: "基于OTel的HTTP链路追踪",
            slug: "otel-http-tracing",
            createdAt: "2024/04/08"
          },
          {
            id: 19,
            title: "Jaeger快速指南",
            slug: "jaeger-quick-guide",
            createdAt: "2024/03/24"
          },
          {
            id: 20,
            title: "OpenTelemetry Go快速指南",
            slug: "opentelemetry-go-quick-guide",
            createdAt: "2024/03/17"
          },
          {
            id: 21,
            title: "OpenTelemetry 介绍",
            slug: "opentelemetry-introduction",
            createdAt: "2024/03/17"
          }
        ];

        // Group notes by year
        const grouped = mockNotes.reduce((acc, note) => {
          const year = note.createdAt.substring(0, 4);

          // Find year group
          let yearGroup = acc.find(group => group.year === year);
          if (!yearGroup) {
            yearGroup = { year, notes: [] };
            acc.push(yearGroup);
          }

          // Add note to year group
          yearGroup.notes.push(note);

          return acc;
        }, [] as ArchiveGroup[]);

        // Sort years in descending order
        grouped.sort((a, b) => parseInt(b.year) - parseInt(a.year));

        // Sort notes by date in descending order
        grouped.forEach(yearGroup => {
          yearGroup.notes.sort((a, b) => {
            return new Date(b.createdAt.replace(/\//g, '-')).getTime() - 
                   new Date(a.createdAt.replace(/\//g, '-')).getTime();
          });
        });

        setArchiveData(grouped);
      } catch (error) {
        console.error('Failed to fetch archives:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchArchives();
  }, []);

  if (loading) {
    return <div className="loading">加载中...</div>;
  }

  return (
    <div className="archives">
      <h1 className="archives-title">文章归档</h1>
      <div className="archives-content">
        {archiveData.map(yearGroup => (
          <div key={yearGroup.year} className="archive-year">
            <h2 className="year-title">{yearGroup.year}</h2>
            <ul className="post-list">
              {yearGroup.notes.map(note => (
                <li key={note.id} className="post-item">
                  <Link to={`/note/${note.id}`}>
                    <span className="post-date">{note.createdAt}</span>
                    <span className="post-link">{note.title}</span>
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Archives;