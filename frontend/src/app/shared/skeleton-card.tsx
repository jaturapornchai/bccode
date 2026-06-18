// Shared skeleton card for loading states.
// Reuse in holding/workspace/menu/login to keep the loading feel consistent.
// Renders N shimmer placeholder cards that match the list-card layout.

export function SkeletonCardList({ count = 3 }: { count?: number }) {
  return (
    <div className="workspace-card-grid">
      {Array.from({ length: count }).map((_, i) => (
        <div className="skeleton-card" aria-hidden="true" key={i}>
          <span className="skeleton-card-avatar" />
          <div className="skeleton-card-lines">
            <span className="skeleton-card-line w-3/4" />
            <span className="skeleton-card-line w-1/2" />
            <span className="skeleton-card-line w-1/3" />
          </div>
        </div>
      ))}
    </div>
  );
}
