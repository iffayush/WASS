type Payload = { repoUrl: string; framework: string; branch: string };
type StartScanButtonProps = {
  payload: Payload;
  onScanComplete: (report: any) => void;
  onScanError: (error: string) => void;
  isLoading: boolean;
  setIsLoading: (loading: boolean) => void;
};

export default function StartScanButton({
  payload,
  onScanComplete,
  onScanError,
  isLoading,
  setIsLoading,
}: StartScanButtonProps) {
  const startScan = async () => {
    if (!payload.repoUrl) {
      onScanError("Repository URL is required");
      return;
    }

    setIsLoading(true);
    try {
      const params = new URLSearchParams({
        repo_url: payload.repoUrl,
        branch: payload.branch,
      });
      
      const response = await fetch(`http://localhost:5678/webhook/security-scan?${params.toString()}`, {
        method: "GET",
      });

      if (!response.ok) {
        throw new Error(`Scan failed: ${response.statusText}`);
      }

      const data = await response.json();
      
      // Parse the nested JSON structure
      const textContent = data.content?.parts?.[0]?.text;
      if (!textContent) {
        throw new Error("Invalid response format");
      }

      // Parse the JSON string within the text
      const report = JSON.parse(textContent);
      onScanComplete(report);
    } catch (error) {
      onScanError(error instanceof Error ? error.message : "Failed to start scan");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <button
      onClick={startScan}
      disabled={isLoading}
      className="w-full py-4 uppercase tracking-[0.12em] bg-red-700 text-[#e5e5e5] border-2 border-[#2a2a2a] hover:bg-red-800 active:bg-red-900 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-150 rounded-none flex items-center justify-center gap-3"
    >
      {isLoading && (
        <div className="w-5 h-5 border-2 border-[#e5e5e5] border-t-transparent animate-spin" />
      )}
      <span>{isLoading ? "Scanning..." : "Start Scan"}</span>
    </button>
  );
}