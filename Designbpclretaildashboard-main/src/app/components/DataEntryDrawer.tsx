import { X } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';
import { useState } from 'react';

interface DataEntryDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: NonFuelData) => void;
}

interface NonFuelData {
  qoc: { volume: string };
  lubricants: { volume: string };
  becafe: { transactions: string; amount: string };
  sbiCards: { transactions: string; amount: string };
  ufill: { transactions: string; amount: string };
}

export function DataEntryDrawer({ isOpen, onClose, onSave }: DataEntryDrawerProps) {
  const [formData, setFormData] = useState<NonFuelData>({
    qoc: { volume: '' },
    lubricants: { volume: '' },
    becafe: { transactions: '', amount: '' },
    sbiCards: { transactions: '', amount: '' },
    ufill: { transactions: '', amount: '' },
  });

  const handleSave = () => {
    onSave(formData);
    onClose();
  };

  const calculateAvgPerTxn = (amount: string, transactions: string) => {
    const amt = parseFloat(amount) || 0;
    const txn = parseFloat(transactions) || 0;
    if (txn === 0) return '0';
    return (amt / txn).toFixed(2);
  };

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/50 z-40"
            onClick={onClose}
          />
          <motion.div
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', damping: 25, stiffness: 200 }}
            className="fixed right-0 top-0 h-full w-[500px] bg-white shadow-2xl z-50 overflow-y-auto"
          >
            <div className="p-6">
              <div className="flex items-start justify-between mb-6">
                <div>
                  <h2 className="text-2xl font-bold text-gray-900">Add Non-Fuel Sales Data</h2>
                  <p className="text-sm text-gray-600 mt-1">April 2026 • Mumbai Central RO</p>
                </div>
                <button
                  onClick={onClose}
                  className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  <X size={24} className="text-gray-600" />
                </button>
              </div>

              <div className="space-y-6">
                <div className="bg-blue-50 rounded-xl p-4 border border-blue-200">
                  <h3 className="text-lg font-bold text-gray-900 mb-4">Volume Inputs</h3>
                  <div className="space-y-4">
                    <div>
                      <label className="text-sm text-gray-700 mb-2 block font-medium">
                        QOC <span className="text-xs text-gray-500">(kL)</span>
                      </label>
                      <input
                        type="number"
                        value={formData.qoc.volume}
                        onChange={(e) => setFormData({ ...formData, qoc: { volume: e.target.value } })}
                        className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#007BC9]"
                        placeholder="Enter volume in kL"
                      />
                    </div>
                    <div>
                      <label className="text-sm text-gray-700 mb-2 block font-medium">
                        Lubricants <span className="text-xs text-gray-500">(kL)</span>
                      </label>
                      <input
                        type="number"
                        value={formData.lubricants.volume}
                        onChange={(e) => setFormData({ ...formData, lubricants: { volume: e.target.value } })}
                        className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#007BC9]"
                        placeholder="Enter volume in kL"
                      />
                    </div>
                  </div>
                </div>

                <div className="bg-yellow-50 rounded-xl p-4 border border-yellow-200">
                  <h3 className="text-lg font-bold text-gray-900 mb-4">Transactions + Amount</h3>
                  <div className="space-y-4">
                    {[
                      { key: 'becafe', label: 'BeCafe' },
                      { key: 'sbiCards', label: 'SBI Cards' },
                      { key: 'ufill', label: 'UFill' },
                    ].map((item) => {
                      const data = formData[item.key as keyof typeof formData] as { transactions: string; amount: string };
                      return (
                        <div key={item.key} className="bg-white rounded-lg p-4 border border-gray-200">
                          <div className="font-medium text-gray-900 mb-3">{item.label}</div>
                          <div className="grid grid-cols-2 gap-3">
                            <div>
                              <label className="text-xs text-gray-600 mb-1 block">Transactions</label>
                              <input
                                type="number"
                                value={data.transactions}
                                onChange={(e) => {
                                  const newData = { ...formData };
                                  (newData[item.key as keyof typeof formData] as any).transactions = e.target.value;
                                  setFormData(newData);
                                }}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#007BC9] text-sm"
                                placeholder="No."
                              />
                            </div>
                            <div>
                              <label className="text-xs text-gray-600 mb-1 block">Amount (₹)</label>
                              <input
                                type="number"
                                value={data.amount}
                                onChange={(e) => {
                                  const newData = { ...formData };
                                  (newData[item.key as keyof typeof formData] as any).amount = e.target.value;
                                  setFormData(newData);
                                }}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#007BC9] text-sm"
                                placeholder="₹"
                              />
                            </div>
                          </div>
                          {data.transactions && data.amount && (
                            <div className="mt-2 text-xs text-gray-600">
                              Avg per txn: ₹{calculateAvgPerTxn(data.amount, data.transactions)}
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={handleSave}
                  className="flex-1 px-6 py-3 bg-[#FFE000] text-[#007BC9] rounded-lg font-bold hover:bg-[#FFD000] transition-colors shadow-md"
                >
                  Save Data
                </button>
                <button
                  onClick={onClose}
                  className="px-6 py-3 border-2 border-gray-300 rounded-lg font-medium hover:bg-gray-50 transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}
