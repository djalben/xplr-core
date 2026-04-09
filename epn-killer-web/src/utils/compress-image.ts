/**
 * Client-side image compression using Canvas API.
 * Resizes to max 1200px width and compresses to WebP (or JPEG fallback).
 * Target: <300KB while maintaining high visual quality.
 */
export async function compressImage(file: File): Promise<File> {
  const MAX_WIDTH = 1200;
  const TARGET_SIZE = 300 * 1024; // 300KB
  const INITIAL_QUALITY = 0.85;
  const MIN_QUALITY = 0.5;

  return new Promise((resolve, reject) => {
    const img = new Image();
    const url = URL.createObjectURL(file);

    img.onload = () => {
      URL.revokeObjectURL(url);

      // Calculate dimensions
      let { width, height } = img;
      if (width > MAX_WIDTH) {
        height = Math.round((height * MAX_WIDTH) / width);
        width = MAX_WIDTH;
      }

      const canvas = document.createElement('canvas');
      canvas.width = width;
      canvas.height = height;
      const ctx = canvas.getContext('2d');
      if (!ctx) {
        reject(new Error('Canvas not supported'));
        return;
      }
      ctx.drawImage(img, 0, 0, width, height);

      // Try WebP first, fallback to JPEG
      const supportsWebP = canvas.toDataURL('image/webp').startsWith('data:image/webp');
      const mimeType = supportsWebP ? 'image/webp' : 'image/jpeg';
      const extension = supportsWebP ? '.webp' : '.jpg';

      // Iterative quality reduction to hit target size
      let quality = INITIAL_QUALITY;

      const tryCompress = () => {
        canvas.toBlob(
          (blob) => {
            if (!blob) {
              reject(new Error('Compression failed'));
              return;
            }

            if (blob.size <= TARGET_SIZE || quality <= MIN_QUALITY) {
              const compressedFile = new File(
                [blob],
                file.name.replace(/\.[^.]+$/, extension),
                { type: mimeType }
              );
              console.log(
                `[compress] ${(file.size / 1024).toFixed(0)}KB → ${(blob.size / 1024).toFixed(0)}KB (q=${quality.toFixed(2)}, ${width}x${height}, ${mimeType})`
              );
              resolve(compressedFile);
            } else {
              quality -= 0.05;
              tryCompress();
            }
          },
          mimeType,
          quality
        );
      };

      tryCompress();
    };

    img.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error('Failed to load image'));
    };

    img.src = url;
  });
}
