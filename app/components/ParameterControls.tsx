import { useState } from "react";
import type { ProcessingParameters } from "~/types";

interface ParameterControlsProps {
  parameters: ProcessingParameters;
  onChange: (params: ProcessingParameters) => void;
}

export function ParameterControls({ parameters, onChange }: ParameterControlsProps) {
  const [activeTab, setActiveTab] = useState<string>('binarization');

  const handleChange = (section: keyof ProcessingParameters, field: string, value: any) => {
    onChange({
      ...parameters,
      [section]: {
        ...parameters[section],
        [field]: value,
      },
    });
  };

  const tabs = [
    { id: 'binarization', label: 'Binarization' },
    { id: 'morphology', label: 'Morphology' },
    { id: 'noise', label: 'Noise' },
    { id: 'scaling', label: 'Scaling' },
  ];

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Tab Navigation */}
      <div className="border-b border-gray-200">
        <nav className="flex -mb-px">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`py-2 px-6 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.id
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="p-6">
        {/* Binarization Tab */}
        {activeTab === 'binarization' && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Method
              </label>
              <select
                value={parameters.binarization.method}
                onChange={(e) => handleChange('binarization', 'method', e.target.value)}
                className="input-field"
              >
                <option value="sauvola">Sauvola</option>
                <option value="niblack">Niblack</option>
                <option value="otsu">Otsu</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                <span className="group relative inline-flex items-center gap-1">
                  Window Size: {parameters.binarization.windowSize}
                  <span className="invisible group-hover:visible absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-xs rounded whitespace-nowrap z-10">
                    Size of the local window for adaptive thresholding (5-51, odd numbers)
                  </span>
                </span>
              </label>
              <input
                type="range"
                min="5"
                max="51"
                step="2"
                value={parameters.binarization.windowSize}
                onChange={(e) => {
                  const val = parseInt(e.target.value);
                  handleChange('binarization', 'windowSize', val % 2 === 0 ? val + 1 : val);
                }}
                className="slider-track"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                <span className="group relative inline-flex items-center gap-1">
                  Threshold K: {parameters.binarization.k.toFixed(2)}
                  <span className="invisible group-hover:visible absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-xs rounded whitespace-nowrap z-10">
                    Controls the threshold bias (0.1-1.0, higher = more white)
                  </span>
                </span>
              </label>
              <input
                type="range"
                min="0.1"
                max="1.0"
                step="0.01"
                value={parameters.binarization.k}
                onChange={(e) => handleChange('binarization', 'k', parseFloat(e.target.value))}
                className="slider-track"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                <span className="group relative inline-flex items-center gap-1">
                  Parameter R: {parameters.binarization.r}
                  <span className="invisible group-hover:visible absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-xs rounded whitespace-nowrap z-10">
                    Dynamic range of standard deviation (0-255)
                  </span>
                </span>
              </label>
              <input
                type="range"
                min="0"
                max="255"
                value={parameters.binarization.r}
                onChange={(e) => handleChange('binarization', 'r', parseInt(e.target.value))}
                className="slider-track"
              />
            </div>
          </div>
        )}

        {/* Morphology Tab */}
        {activeTab === 'morphology' && (
          <div className="space-y-4">
            <div>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={parameters.morphology.enabled}
                  onChange={(e) => handleChange('morphology', 'enabled', e.target.checked)}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <span className="text-sm font-medium text-gray-700">Enable Morphological Operations</span>
              </label>
            </div>

            {parameters.morphology.enabled && (
              <>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Operation
                  </label>
                  <select
                    value={parameters.morphology.operation}
                    onChange={(e) => handleChange('morphology', 'operation', e.target.value)}
                    className="input-field"
                  >
                    <option value="close">Close (fill gaps)</option>
                    <option value="open">Open (remove noise)</option>
                    <option value="dilate">Dilate (thicken)</option>
                    <option value="erode">Erode (thin)</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    <span className="group relative inline-flex items-center gap-1">
                      Kernel Size: {parameters.morphology.kernelSize}
                      <span className="invisible group-hover:visible absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-xs rounded whitespace-nowrap z-10">
                        Size of the morphological kernel (3-9, odd numbers)
                      </span>
                    </span>
                  </label>
                  <input
                    type="range"
                    min="3"
                    max="9"
                    step="2"
                    value={parameters.morphology.kernelSize}
                    onChange={(e) => {
                      const val = parseInt(e.target.value);
                      handleChange('morphology', 'kernelSize', val % 2 === 0 ? val + 1 : val);
                    }}
                    className="slider-track"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    <span className="group relative inline-flex items-center gap-1">
                      Iterations: {parameters.morphology.iterations}
                      <span className="invisible group-hover:visible absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-xs rounded whitespace-nowrap z-10">
                        Number of times to apply the operation (1-3)
                      </span>
                    </span>
                  </label>
                  <input
                    type="range"
                    min="1"
                    max="3"
                    value={parameters.morphology.iterations}
                    onChange={(e) => handleChange('morphology', 'iterations', parseInt(e.target.value))}
                    className="slider-track"
                  />
                </div>
              </>
            )}
          </div>
        )}

        {/* Noise Tab */}
        {activeTab === 'noise' && (
          <div className="space-y-4">
            <div>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={parameters.noise.enabled}
                  onChange={(e) => handleChange('noise', 'enabled', e.target.checked)}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <span className="text-sm font-medium text-gray-700">Enable Noise Reduction</span>
              </label>
            </div>

            {parameters.noise.enabled && (
              <>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Method
                  </label>
                  <select
                    value={parameters.noise.method}
                    onChange={(e) => handleChange('noise', 'method', e.target.value)}
                    className="input-field"
                  >
                    <option value="binary">Binary (isolated pixels)</option>
                    <option value="median">Median Filter</option>
                  </select>
                </div>

                {parameters.noise.method === 'binary' && (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      <span className="group relative inline-flex items-center gap-1">
                        Threshold: {parameters.noise.threshold}
                        <span className="invisible group-hover:visible absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-xs rounded whitespace-nowrap z-10">
                          Minimum neighbors with same value (1-8)
                        </span>
                      </span>
                    </label>
                    <input
                      type="range"
                      min="1"
                      max="8"
                      value={parameters.noise.threshold}
                      onChange={(e) => handleChange('noise', 'threshold', parseInt(e.target.value))}
                      className="slider-track"
                    />
                  </div>
                )}

                {parameters.noise.method === 'median' && (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      <span className="group relative inline-flex items-center gap-1">
                        Window Size: {parameters.noise.windowSize}
                        <span className="invisible group-hover:visible absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-xs rounded whitespace-nowrap z-10">
                          Size of the median filter window (3-7, odd numbers)
                        </span>
                      </span>
                    </label>
                    <input
                      type="range"
                      min="3"
                      max="7"
                      step="2"
                      value={parameters.noise.windowSize}
                      onChange={(e) => {
                        const val = parseInt(e.target.value);
                        handleChange('noise', 'windowSize', val % 2 === 0 ? val + 1 : val);
                      }}
                      className="slider-track"
                    />
                  </div>
                )}
              </>
            )}
          </div>
        )}

        {/* Scaling Tab */}
        {activeTab === 'scaling' && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Scaling Method
              </label>
              <select
                value={parameters.scaling.method}
                onChange={(e) => handleChange('scaling', 'method', e.target.value)}
                className="input-field"
              >
                <option value="none">None</option>
                <option value="2x">2x Scale</option>
                <option value="3x">3x Scale</option>
                <option value="4x">4x Scale</option>
              </select>
            </div>

            {parameters.scaling.method !== 'none' && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Algorithm
                </label>
                <select
                  value={parameters.scaling.algorithm}
                  onChange={(e) => handleChange('scaling', 'algorithm', e.target.value)}
                  className="input-field"
                >
                  <option value="scale2x">Scale2x (pixel art)</option>
                  <option value="scale3x">Scale3x (pixel art)</option>
                  <option value="nearest">Nearest Neighbor</option>
                  <option value="bilinear">Bilinear</option>
                </select>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}