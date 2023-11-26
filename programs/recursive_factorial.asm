start:
    jmp main

factorial:
    cmp ax, 1
    jle end ; base case

    push ax
    sub ax, 1
    call factorial

    pop bx
    call mult_ab
    ret

end:
    ret

; mul ax, bx (mul unimplemented)
mult_ab:
    mov cx, ax
    mov dx, bx
    sub dx, 1
    mult_loop:
        add ax, cx
        sub dx, 1
        jnz mult_loop
    ret

main:
      mov ax, 5
      call factorial