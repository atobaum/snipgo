interface SnippetMetaInfoProps {
  createdAt: string;
  updatedAt: string;
}

export function SnippetMetaInfo({ createdAt, updatedAt }: SnippetMetaInfoProps) {
  return (
    <>
      <div>
        <span className="text-gray-500 text-xs">
          Created: {new Date(createdAt).toLocaleDateString()}
        </span>
      </div>
      <div>
        <span className="text-gray-500 text-xs">
          Updated: {new Date(updatedAt).toLocaleDateString()}
        </span>
      </div>
    </>
  );
}
