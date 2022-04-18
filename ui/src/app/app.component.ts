import { Component, OnInit } from '@angular/core';
import { Title } from '@angular/platform-browser';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
  title = 'PWA Icon Generator';

  constructor(private ts: Title) {}

  ngOnInit(): void {
    this.ts.setTitle(this.title);
  }
}
