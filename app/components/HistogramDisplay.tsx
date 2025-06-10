import { BarChart3 } from "lucide-react";

interface HistogramDisplayProps {
  data: number[];
}

export function HistogramDisplay({ data }: HistogramDisplayProps) {
  if (!data || data.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow p-4">
        <h3 className="text-sm font-semibold mb-2 flex items-center gap-2">
          <BarChart3 className="w-4 h-4 text-blue-600" />
          Histogram
        </h3>
        <div className="h-32 flex items-center justify-center text-gray-400 text-sm">
          No histogram data available
        </div>
      </div>
    );
  }

  // Normalize histogram data
  const maxValue = Math.max(...data);
  const normalizedData = data.map(value => (value / maxValue) * 100);

  // Sample every nth value for display (showing all 256 values would be too dense)
  const sampleRate = 4;
  const sampledData = normalizedData.filter((_, index) => index % sampleRate === 0);

  return (
    <div className="bg-white rounded-lg shadow p-4">
      <h3 className="text-sm font-semibold mb-2 flex items-center gap-2">
        <BarChart3 className="w-4 h-4 text-blue-600" />
        Histogram
      </h3>
      <div className="h-32 flex items-end gap-px">
        {sampledData.map((value, index) => (
          <div
            key={index}
            className="flex-1 bg-blue-400 hover:bg-blue-500 transition-colors"
            style={{ height: `${value}%` }}
            title={`Level ${index * sampleRate}: ${Math.round(data[index * sampleRate])}`}
          />
        ))}
      </div>
      <div className="flex justify-between mt-1 text-xs text-gray-500">
        <span>0</span>
        <span>255</span>
      </div>
    </div>
  );
}