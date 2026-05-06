import { X } from 'lucide-react';
import { useState } from 'react';

interface TargetValues {
  [key: string]: number;
}

export function SetTargetsModal({ onClose, onSave, initialTargets }: { onClose: () => void; onSave: (targets: any) => void; initialTargets: any }) {
  const [fuelTargets, setFuelTargets] = useState<TargetValues>({
    MS: initialTargets.fuel.MS || 2415,
    HSD: initialTargets.fuel.HSD || 4098,
    SPEED: initialTargets.fuel.SPEED || 336,
  });

  const [nonFuelTargets, setNonFuelTargets] = useState<TargetValues>({
    QOC: initialTargets.nonFuel.QOC || 793,
    Lubricants: initialTargets.nonFuel.Lubricants || 463,
    UFill: initialTargets.nonFuel.UFill || 341,
    SBI: initialTargets.nonFuel.SBI || 183,
    BeCafe: initialTargets.nonFuel.BeCafe || 152,
  });

  const handleFuelChange = (product: string, value: string) => {
    setFuelTargets({ ...fuelTargets, [product]: parseFloat(value) || 0 });
  };

  const handleNonFuelChange = (product: string, value: string) => {
    setNonFuelTargets({ ...nonFuelTargets, [product]: parseFloat(value) || 0 });
  };

  const fuelTotal = Object.values(fuelTargets).reduce((sum, val) => sum + val, 0);
  const nonFuelTotal = Object.values(nonFuelTargets).reduce((sum, val) => sum + val, 0);
  const grandTotal = fuelTotal + nonFuelTotal;

  const handleSave = () => {
    onSave({
      fuel: fuelTargets,
      nonFuel: nonFuelTargets,
    });
    onClose();
  };

  const numberBasedProducts = ['UFill', 'SBI', 'BeCafe'];

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
          <h2 className="text-xl text-gray-900">Set Targets</h2>
          <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-lg transition-colors">
            <X size={20} className="text-gray-600" />
          </button>
        </div>

        <div className="p-6 space-y-6">
          <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
            <div className="px-6 py-3" style={{ backgroundColor: '#007BC9' }}>
              <h3 className="text-white">Fuel Targets (kL)</h3>
            </div>
            <div className="p-6 space-y-4">
              {Object.keys(fuelTargets).map((product) => (
                <div key={product} className="flex items-center justify-between">
                  <label className="text-gray-700 w-32">{product}</label>
                  <div className="flex items-center gap-2">
                    <input
                      type="number"
                      value={fuelTargets[product]}
                      onChange={(e) => handleFuelChange(product, e.target.value)}
                      className="w-48 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2"
                      onFocus={(e) => e.target.style.boxShadow = '0 0 0 2px rgba(0, 123, 201, 0.2)'}
                      onBlur={(e) => e.target.style.boxShadow = ''}
                    />
                    <span className="text-gray-600">kL</span>
                  </div>
                </div>
              ))}
              <div className="border-t border-gray-200 pt-4 flex items-center justify-between">
                <label className="text-gray-900 font-semibold">Fuel Total</label>
                <div className="text-xl text-gray-900 font-semibold">{fuelTotal.toLocaleString()} kL</div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
            <div className="px-6 py-3" style={{ backgroundColor: '#007BC9' }}>
              <h3 className="text-white">Non-Fuel Targets</h3>
            </div>
            <div className="p-6 space-y-4">
              {Object.keys(nonFuelTargets).map((product) => {
                const isNumberBased = numberBasedProducts.includes(product);
                return (
                  <div key={product} className="flex items-center justify-between">
                    <label className="text-gray-700 w-32">{product}</label>
                    <div className="flex items-center gap-2">
                      <input
                        type="number"
                        value={nonFuelTargets[product]}
                        onChange={(e) => handleNonFuelChange(product, e.target.value)}
                        className="w-48 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2"
                        onFocus={(e) => e.target.style.boxShadow = '0 0 0 2px rgba(0, 123, 201, 0.2)'}
                        onBlur={(e) => e.target.style.boxShadow = ''}
                      />
                      <span className="text-gray-600 w-16">{isNumberBased ? 'Nos' : 'kL'}</span>
                    </div>
                  </div>
                );
              })}
              <div className="border-t border-gray-200 pt-4 flex items-center justify-between">
                <label className="text-gray-900 font-semibold">Non-Fuel Total</label>
                <div className="text-xl text-gray-900 font-semibold">{nonFuelTotal.toLocaleString()}</div>
              </div>
            </div>
          </div>

          <div className="bg-gray-50 rounded-lg p-6 border-2" style={{ borderColor: '#FFE000' }}>
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-gray-900 font-semibold mb-1">Total Sales Target</h3>
                <p className="text-sm text-gray-600">Combined fuel and non-fuel targets</p>
              </div>
              <div className="text-3xl text-gray-900 font-bold">{grandTotal.toLocaleString()}</div>
            </div>
          </div>
        </div>

        <div className="sticky bottom-0 bg-white border-t border-gray-200 px-6 py-4 flex items-center justify-end gap-3">
          <button
            onClick={onClose}
            className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            className="px-6 py-2 rounded-lg text-gray-900 transition-all hover:shadow-md"
            style={{ backgroundColor: '#FFE000' }}
          >
            Save Targets
          </button>
        </div>
      </div>
    </div>
  );
}
