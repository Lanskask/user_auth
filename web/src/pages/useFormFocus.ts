import { useEffect, useRef } from "react";

/** After a failed submit, moves focus to the first invalid field so keyboard and screen-reader users land on it. */
export function useFocusFirstInvalid(attempt: number) {
  const formRef = useRef<HTMLFormElement>(null);
  useEffect(() => {
    if (attempt > 0) formRef.current?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
  }, [attempt]);
  return formRef;
}
