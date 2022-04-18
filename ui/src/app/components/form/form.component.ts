import { HttpClient } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { delay, from, Observable, of } from 'rxjs';
import { environment } from 'src/environments/environment';
import { UUID } from 'angular2-uuid';

@Component({
  selector: 'app-form',
  templateUrl: './form.component.html',
  styleUrls: ['./form.component.css']
})
export class FormComponent implements OnInit {

  apigw: string;
  uniqueName: string = '';
  fileSelected: boolean = false;
  generating: boolean = false;
  dwnldReady: boolean = false;
  private file: File | null = null;

  constructor(private http: HttpClient) {
    this.apigw = environment.apigw;
  }

  ngOnInit(): void {
  }

  upload(event: Event): void {
    const target = event.target as HTMLInputElement;
    const tmpfile: File = (target.files as FileList)[0];
    this.file = tmpfile;
    this.fileSelected = true;
  }

  pushFile(): void {
    if (this.file === null) {
      this.fileSelected = false;
      return;
    }
    this.uniqueName = UUID.UUID();
    this.apigw += this.uniqueName;
    this.generating = true;
    const putResponseObservable: Observable<Object> = this.http.put(this.apigw, this.file);
    putResponseObservable.subscribe({
      next: (res: Object) => {
        console.log('put response: ' + JSON.stringify(res));
        // Load link to output zip
        const dummyObservable = of(res);
        dummyObservable.pipe(delay(2500));
        dummyObservable.subscribe({
          next: () => {
            this.generating = false;
            this.dwnldReady = true;
          }
        });
      },
      error: (err: any) => {
        console.log('error: ' + JSON.stringify(err));
        this.generating = false;
      },
      complete: () => {
        console.log('put response complete.');
      }
    });
    this.fileSelected = false;
    this.file = null;
  }

  downloadIcons(): void {
    const zipUrl = environment.s3out + this.uniqueName + '.icons.zip';
    window.open(zipUrl, '_blank');
  }

}
