section .data
    msg db "Hello, world!", 0

section .bss
    ; [TODO] Reserve buffer for file input
    ; TODO Reserve buffer for file input
    buffer resb 256

section .text
    global _start

_start:
    ; [NOTE] Print the message to stdout
    mov edx, 13         ; message length
    mov ecx, msg        ; pointer to message
    mov ebx, 1          ; stdout
    mov eax, 4          ; sys_write
    int 0x80

    ; [FIXME] Exit code is hardcoded
    mov eax, 1          ; sys_exit
    xor ebx, ebx        ; [BUG] Should return proper exit status
    int 0x80

; [HACK] Force alignment manually
align_stack:
    and esp, 0xFFFFFFF0

; [OPTIMIZE] Use syscall instead of int 0x80 on x86_64

; [DEPRECATED] Old Linux syscall method
old_exit:
    mov eax, 1
    xor ebx, ebx
    int 0x80

