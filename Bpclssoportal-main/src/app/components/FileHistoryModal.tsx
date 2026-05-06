import { X, FileText, Download, Trash2 } from 'lucide-react';
import { format } from 'date-fns';

interface FileHistoryItem {
  id: string;
  name: string;
  uploadDate: Date;
  size: string;
}

export function FileHistoryModal({
  onClose,
  history,
  onDelete
}: {
  onClose: () => void;
  history: FileHistoryItem[];
  onDelete: (id: string) => void;
}) {
  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-3xl max-h-[80vh] overflow-y-auto">
        <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
          <div>
            <h2 className="text-xl text-gray-900">File Upload History</h2>
            <p className="text-sm text-gray-600 mt-1">Recently uploaded files</p>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-lg transition-colors">
            <X size={20} className="text-gray-600" />
          </button>
        </div>

        <div className="p-6">
          {history.length === 0 ? (
            <div className="text-center py-12">
              <FileText size={48} className="mx-auto text-gray-400 mb-3" />
              <h3 className="text-gray-900 mb-2">No files uploaded yet</h3>
              <p className="text-gray-600 text-sm">Upload Excel files to see them here</p>
            </div>
          ) : (
            <div className="space-y-3">
              {history.map((file) => (
                <div
                  key={file.id}
                  className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
                >
                  <div className="flex items-center gap-3 flex-1">
                    <div className="w-10 h-10 rounded-lg flex items-center justify-center" style={{ backgroundColor: '#007BC9' }}>
                      <FileText size={20} className="text-white" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <h4 className="text-gray-900 truncate">{file.name}</h4>
                      <div className="flex items-center gap-3 text-sm text-gray-600 mt-1">
                        <span>{format(file.uploadDate, 'dd MMM yyyy, HH:mm')}</span>
                        <span>•</span>
                        <span>{file.size}</span>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => onDelete(file.id)}
                      className="p-2 hover:bg-red-50 rounded-lg transition-colors group"
                      title="Delete"
                    >
                      <Trash2 size={18} className="text-gray-400 group-hover:text-red-600" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="sticky bottom-0 bg-white border-t border-gray-200 px-6 py-4 flex items-center justify-between">
          <p className="text-sm text-gray-600">{history.length} file{history.length !== 1 ? 's' : ''} in history</p>
          <button
            onClick={onClose}
            className="px-6 py-2 rounded-lg text-gray-900 transition-all hover:shadow-md"
            style={{ backgroundColor: '#FFE000' }}
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
