import type { FC } from "react";
import styles from "./Form.module.css";

type FormType = "registration" | "authorization"

export interface FormProps {
    formType?: FormType
}

export const Form: FC<FormProps> = ({ formType }) => {
    if (!formType || (formType === "registration")) {
        return (
            <form className={styles.form}>
                <h1
                    className={`${styles.h1}`}
                >Registration</h1>
                <label className={`${styles.input_label}`} htmlFor="username">username</label>
                <input
                    pattern="[0-9a-zA-z]{3,8}"
                    id="username"
                    className={styles.input}
                    inputMode="text" />
                <label className={`${styles.input_label}`} htmlFor="email">email</label>
                <input
                    className={`${styles.input}`}
                    inputMode="email"
                    id="email" />
                <label className={`${styles.input_label}`} htmlFor="password">password</label>
                <input
                    className={`${styles.input}`}
                    type="password"
                    id="password" />
                <div className={`${styles.button_zone}`}>
                    <button type="submit"
                        onClick={(e) => {
                            e.preventDefault();


                        }}
                    >OK</button>
                </div>
            </form>
        )
    }
} 